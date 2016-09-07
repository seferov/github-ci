#!/bin/bash

DIR=$1

# Get the latest version
git --git-dir="${DIR}/.git" --work-tree="${DIR}" pull

# Update database
$DIR"/app/console" doctrine:migrations:migrate --env=prod --no-interaction
