#!/bin/bash

openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -sha256 -days 3650 -nodes -subj "/C=US/ST=FL/L=Gainesville/O=FrontRunner/OU=/CN=localhost"

# Change the STOREFRONT_KEY setting:
storefrontKey=$(openssl rand -base64 32)
sed -i -E "s|(STOREFRONT_KEY[ ]*=).*|\1 \"$storefrontKey\"|" .env
printf "Setting STOREFRONT_KEY to: $storefrontKey"