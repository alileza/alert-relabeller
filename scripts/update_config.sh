#!/bin/bash

# Define the URL where you want to send the POST request
URL="http://localhost:9999/config"

# Define the path to the YAML file you want to send
YAML_FILE="config.yml"  # Replace with the actual path to your YAML file

# Send the POST request with curl
curl -v -X POST \
     -H "Content-Type: application/yaml" \
     --data-binary "@$YAML_FILE" \
     "$URL"
