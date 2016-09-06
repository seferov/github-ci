#!/bin/bash

DIR=$1

# Get the latest version
git --git-dir="${DIR}/.git" --work-tree="${DIR}" pull

chmod -R 777 "${DIR}/app/cache/"
chmod -R 777 "${DIR}/app/logs/"
