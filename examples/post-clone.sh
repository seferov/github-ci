#!/bin/bash

DIR=$1

# Install composer
/usr/local/bin/composer install --prefer-dist --no-interaction --optimize-autoloader -v --working-dir=$DIR

# Create database

#$DIR"/app/console" doctrine:migrations:migrate --env=prod --no-interaction

chown www-data:www-data -R $DIR
chmod 777 -R $DIR"/app/cache/"
chmod 777 -R $DIR"/app/logs/"
