#!/bin/bash

# ++++++++++++++++++++ START ELASTICSEARCH AND LLAMA ++++++++++++++++++++
docker compose -f ./docker/docker-compose-elasticollama.yml up -d elasticsearch 2>/dev/null
docker compose -f ./docker/docker-compose-elasticollama.yml up -d ollama 2>/dev/null


# ++++++++++++++++++++ CREATE ELASTICSEARCH INDEX ++++++++++++++++++++
# wait until the elasticsearch is up and running
while true; do
  curl -s http://localhost:9200 > /dev/null
  if [ $? -eq 0 ]; then
    break
  fi
  sleep 1
done

echo "Ingesting knowledge base..."
curl -u elastic:$ELASTICSEARCH_PASSWORD -X PUT "http://localhost:9200/knowledge_base" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "policy_name": { "type": "text" },
      "criteria": {
        "properties": {
          "response_time_ms": { "type": "float" },
          "error_rate": { "type": "float" },
          "vulnerability_count": { "type": "integer" },
          "intrusion_attempts": { "type": "integer" }
        }
      },
      "recommended_actions": {
        "properties": {
          "switch_language": { "type": "keyword" },
          "switch_format": { "type": "keyword" },
          "switch_os": { "type": "keyword" },
          "rotate_ip": { "type": "boolean" }
        }
      }
    }
  }
}
'

# Verify index creation
# curl -u elastic:changeme "http://localhost:9200/_cat/indices?v"
check=`curl -u elastic:$ELASTICSEARCH_PASSWORD "http://localhost:9200/_cat/indices?v" | grep knowledge_base`

if [ -z "$check" ]; then
  echo "Failed to create index"
  exit 1
fi

# ++++++++++++++++++++ INGEST KNOWLEDGE DATA ++++++++++++++++++++

# Check if Python is installed
if ! command -v python3 &> /dev/null
then
    echo "Python3 could not be found. Please install Python3 to proceed."
    exit
fi

# Read JSON data and format for bulk API
BULK_DATA=$(python3 <<EOF
import json
import sys

with open("$KNOWLEDGE_DATA", "r") as f:
    data = json.load(f)

bulk = ""
for doc in data:
    action = { "index": { "_index": "$ELASTICSEARCH_INDEX" } }
    bulk += json.dumps(action) + "\n"
    bulk += json.dumps(doc) + "\n"

print(bulk)
EOF
)

# Send bulk data to Elasticsearch
RESPONSE=$(echo "$BULK_DATA" | curl -u $ELASTICSEARCH_USER:$ELASTICSEARCH_PASSWORD -X POST "$ELASTICSEARCH_URL/_bulk" -H 'Content-Type: application/json' --data-binary @-)

# Check for errors
if echo "$RESPONSE" | grep -q '"errors":true'; then
    echo "Errors occurred during bulk ingestion:"
    echo "$RESPONSE" | python3 -m json.tool
else
    echo "Knowledge data ingested successfully."
fi

## TEST ELASTICSEARCH: 
# curl -u elastic:password "http://localhost:9200/knowledge_base/_search?pretty"

# TEST OLLAMA
# curl http://localhost:11434/api/chat -d '{
#   "model": "llama3:latest",
#   "messages": [
#     { "role": "user", "content": "Return only the word pong if you are able to answer" }
#   ]
# }' | grep pong 2>/dev/null

echo "Execute the following command to download the model:"
echo "docker exec -it ollama ollama pull llama3:latest"
