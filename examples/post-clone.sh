#!/bin/bash

DIR=$1
NAME=$2
DBNAME=$NAME

echo $NAME

# Install composer
echo "Installing vendor files by composer"
/usr/local/bin/composer install --prefer-dist --no-interaction --optimize-autoloader -v --working-dir=$DIR

# Fix user permissions
chown www-data:www-data -R $DIR
chmod 777 -R $DIR"/app/cache/"
chmod 777 -R $DIR"/app/logs/"
