# LoRaWAN Mapper

## Usage

```
$ ./lorawan-mapper -h 

Capture and visualize LoRaWAN messages

Usage:
  lorawan-mapper [flags]

Flags:
  -b, --bind string   Bind address (ip:port) (default "0.0.0.0:1700")
  -h, --help          help for lorawan-mapper
  -l, --log string    Log File Path (default "gateway.log")
```

## Neo4j Database

Neo4j Docker container:

```
docker run \
    --name testneo4j \
    -p7474:7474 -p7687:7687 \
    -d \
    -v $HOME/neo4j/data:/data \
    -v $HOME/neo4j/logs:/logs \
    -v $HOME/neo4j/import:/var/lib/neo4j/import \
    -v $HOME/neo4j/plugins:/plugins \
    --env NEO4J_AUTH=neo4j/password \
    neo4j:latest
```

## Neo4j: Sample Queries

Visualize relationships between End-Devices and Join Servers (OTAA procedure)

```
MATCH (d:Device)-[:OTAA]->(j:JoinServer) RETURN d,j
```

Visualize relationships between End-Devices and Networks (data up)

```
MATCH (d:Device)-[:DataUp]->(n:Network) RETURN d,n
```

## Neo4j: Clear Database

```
DROP CONSTRAINT join_eui
DROP CONSTRAINT dev_eui
DROP CONSTRAINT nwk_addr
DROP CONSTRAINT nwk_id
MATCH (d:Device)-[r:OTAA]->() DELETE r
MATCH (d:Device)-[r:DataUp]->() DELETE r
MATCH (d:Device) DELETE d
MATCH (j:JoinServer) DELETE j
MATCH (n:Network) DELETE n
```