#!/bin/bash

cleanUp() {
    rm -rf $1
    exit $2
}

TARGET="resources/public/css"

if [ -d ${TARGET}/line-awesome/ ]; then
    echo "Line awesome already exists"
    exit 0
fi

TMPLOC=$(mktemp -d)
MD5="2a9ec1fd85e09bb7b3f277ad64aebb43 ${TMPLOC}/line-awesome.zip"

LA="${TMPLOC}/line-awesome.zip"
URL="https://maxcdn.icons8.com/fonts/line-awesome/1.1/line-awesome.zip"

echo "Now downloading line awesome"
wget -O ${LA} ${URL}

echo "Checking line awesome"
echo ${MD5} | md5sum --status -c
CHECKSUM=$?

if [ ! $CHECKSUM ]; then
    echo "Line awesome did not match md5sum exiting ..."
    cleanUp ${TMPLOC} $CHECKSUM
fi

# install
echo "Unzipping"
unzip ${LA} -d ${TARGET}

# clean up
cleanUp ${TMPLOC} 0
