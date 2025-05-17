# LoRaWAN Mapper

A tool for capturing LoRaWAN messages and visualizing networks.

## Build

Clone the repository and build the project with Golang:

```
$ git clone https://github.com/pasknel/lorawan-mapper.git
$ cd lorawan-mapper
$ go build
```

## Usage

```
$ ./lorawan-mapper -h 

Capture and visualize LoRaWAN messages

Usage:
  lorawan-mapper [flags]
  lorawan-mapper [command]

Available Commands:
  capture     Capture LoRaWAN messages
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  setup       Setup neo4j database

Flags:
  -h, --help   help for lorawan-mapper

Use "lorawan-mapper [command] --help" for more information about a command.
```

## Neo4j Database

Start a Neo4j Docker container:

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

Create a config.yaml file with the credentials for Neo4j:

```yaml
NEO4J_URI: neo4j://localhost:7687
NEO4J_USERNAME: neo4j
NEO4J_PASSWORD: password
```

Setup the Neo4j database:

```
$ ./lorawan-mapper setup
```

## Capturing Packets

Start the gateway listener (runs by default on 0.0.0.0:1700)

```
$ ./lorawan-mapper capture
```

LoRaWAN-Mapper creates two files:
- A PCAP file (default: `capture.pcap`)
- A named pipe (default: `/tmp/lorawan`)

Open the named pipe in wireshark:

```
$ sudo wireshark -k -i /tmp/lorawan
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