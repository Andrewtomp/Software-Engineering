#!/bin/bash

openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -sha256 -days 3650 -nodes -subj "/C=US/ST=FL/L=Gainesville/O=FrontRunner/OU=/CN=localhost"

sessionAuthKey=$(openssl rand -base64 32)
storefrontKey=$(openssl rand -base64 32)

if [[ "$(uname)" == "Darwin" ]]
then
    sed -i '' -E "s|(SESSION_AUTH_KEY[ ]*=).*|\1 \"$sessionAuthKey\"|" .env
    sed -i '' -E "s|(STOREFRONT_KEY[ ]*=).*|\1 \"$storefrontKey\"|" .env
elif [[ "$(uname)" == "Linux" ]]
then
    sed -i -E "s|(SESSION_AUTH_KEY[ ]*=).*|\1 \"$sessionAuthKey\"|" .env
    sed -i -E "s|(STOREFRONT_KEY[ ]*=).*|\1 \"$storefrontKey\"|" .env
fi
# # Change the SESSION_AUTH_KEY
# sessionAuthKey=$(openssl rand -base64 32)
# sed -i -E "s|(SESSION_AUTH_KEY[ ]*=).*|\1 \"$sessionAuthKey\"|" .env
# printf "Setting SESSION_AUTH_KEY to: $sessionAuthKey"

# # Change the STOREFRONT_KEY setting:
# storefrontKey=$(openssl rand -base64 32)
# sed -i -E "s|(STOREFRONT_KEY[ ]*=).*|\1 \"$storefrontKey\"|" .env
# printf "Setting STOREFRONT_KEY to: $storefrontKey"