package util

import (
	"encoding/hex"
	"fmt"

	"github.com/brocaar/lorawan"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

type Neo4jClient struct {
	driver neo4j.Driver
}

func NewNeo4jClient() (Neo4jClient, error) {
	viper.SetConfigFile("config.yaml")
	viper.ReadInConfig()

	uri := viper.GetString("NEO4J_URI")
	username := viper.GetString("NEO4J_USERNAME")
	password := viper.GetString("NEO4J_PASSWORD")

	client := Neo4jClient{}

	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""), func(c *neo4j.Config) {
		c.Encrypted = false
	})

	if err != nil {
		return client, err
	}

	client.driver = driver

	return client, nil
}

func (n *Neo4jClient) Close() {
	n.driver.Close()
}

func (n *Neo4jClient) NewSession() (neo4j.Session, error) {
	return n.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
}

func (n *Neo4jClient) Write(query string, parameters map[string]interface{}) error {
	session, err := n.NewSession()
	if err != nil {
		return fmt.Errorf("error creating new neo4j session - err: %v", err)
	}
	defer session.Close()

	_, err = session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(query, parameters)
		if err != nil {
			return nil, fmt.Errorf("error running write transaction - err: %v", err)
		}

		return nil, result.Err()
	})

	if err != nil {
		return err
	}

	return nil
}

func SetupNeo4j() error {
	neo4jClient, err := NewNeo4jClient()
	if err != nil {
		return fmt.Errorf("error creating neo4j client - err: %v", err)
	}
	defer neo4jClient.Close()

	queries := []string{
		"DROP CONSTRAINT join_eui",
		"DROP CONSTRAINT dev_eui",
		"DROP CONSTRAINT nwk_addr",
		"DROP CONSTRAINT nwk_id",
		"CREATE CONSTRAINT join_eui FOR (j:JoinServer) REQUIRE j.JoinEUI IS UNIQUE",
		"CREATE CONSTRAINT dev_eui FOR (d:Device) REQUIRE d.DevEUI IS UNIQUE",
		"CREATE CONSTRAINT nwk_addr FOR (d:Device) REQUIRE d.NwkAddr IS UNIQUE",
		"CREATE CONSTRAINT nwk_id FOR (n:Network) REQUIRE n.NwkId IS UNIQUE",
		"MATCH (d:Device)-[r:OTAA]->(j:JoinServer) DELETE r",
		"MATCH (d:Device)-[r:DataUp]->(n:Network) DELETE r",
		"MATCH (d:Device) DELETE d",
		"MATCH (j:JoinServer) DELETE j",
		"MATCH (n:Network) DELETE n",
	}

	parameters := map[string]interface{}{}

	for _, query := range queries {
		if err := neo4jClient.Write(query, parameters); err != nil {
			log.Errorf("error during neo4j setup - err: %v", err)
		}
	}

	return nil
}

func GraphJoinRequest(joinReq lorawan.JoinRequestPayload) error {
	neo4jClient, err := NewNeo4jClient()
	if err != nil {
		return fmt.Errorf("error creating neo4j client - err: %v", err)
	}
	defer neo4jClient.Close()

	query := "CREATE (j:JoinServer {JoinEUI: $JoinEUI})"
	parameters := map[string]interface{}{"JoinEUI": joinReq.JoinEUI.String()}
	neo4jClient.Write(query, parameters)

	query = "CREATE (d:Device {DevEUI: $DevEUI})"
	parameters = map[string]interface{}{"DevEUI": joinReq.DevEUI.String()}
	neo4jClient.Write(query, parameters)

	query = "MATCH (d:Device {DevEUI: $DevEUI}) " +
		"MATCH (j:JoinServer {JoinEUI: $JoinEUI}) " +
		"MERGE (d)-[:OTAA]->(j)"
	parameters = map[string]interface{}{
		"DevEUI":  joinReq.DevEUI.String(),
		"JoinEUI": joinReq.JoinEUI.String(),
	}
	neo4jClient.Write(query, parameters)

	return nil
}

func GraphDataUp(mac lorawan.MACPayload) error {
	neo4jClient, err := NewNeo4jClient()
	if err != nil {
		return fmt.Errorf("error creating neo4j client - err: %v", err)
	}
	defer neo4jClient.Close()

	nwkAddr := mac.FHDR.DevAddr.String()
	nwkId := hex.EncodeToString(mac.FHDR.DevAddr.NwkID())

	query := "CREATE (d:Device {NwkAddr: $NwkAddr, NwkId: $NwkId})"
	parameters := map[string]interface{}{
		"NwkAddr": nwkAddr,
		"NwkId":   nwkId,
	}
	neo4jClient.Write(query, parameters)

	query = "CREATE (n:Network {NwkId: $NwkId})"
	parameters = map[string]interface{}{
		"NwkId": nwkId,
	}
	neo4jClient.Write(query, parameters)

	query = "MATCH (d:Device {NwkAddr: $NwkAddr, NwkId: $NwkId}) " +
		"MATCH (n:Network {NwkId: $NwkId}) " +
		"MERGE (d)-[:DataUp{fport: $FPort}]->(n)"

	parameters = map[string]interface{}{
		"NwkAddr": nwkAddr,
		"NwkId":   nwkId,
		"FPort":   mac.FPort,
	}

	neo4jClient.Write(query, parameters)

	return nil
}
