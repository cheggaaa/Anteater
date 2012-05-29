#!/bin/bash

############################################
###                                      ###
### Script for migrate image             ###
### from file system to anteater storage ###
###                                      ###
############################################

### Vars
WDIR="/tmp/`basename $0`"
FILE_LIST="$WDIR/file_list"
FILE_LIST_PREFIX="$WDIR/file_list_"
FLOW="20"
DATA_PATH="$1"
UPLOAD_URL="$2"

### Check args
if [ -z "$DATA_PATH" ] || ! [ -d "$DATA_PATH" ] || [ -z "$UPLOAD_URL" ]; then
	cat <<EOF
Usage: $0 data_path upload_url

Example:
	$0 /tmp/imgdir/ http://example.com:8081
EOF
	exit 1
fi

### Action
## Create work dir
install -d $WDIR
## Get file list
echo "Get file lists"
find $DATA_PATH -type f > $FILE_LIST
COUNT_LINES=`wc -l $FILE_LIST | awk '{print $1}'`
STEP=$((COUNT_LINES / FLOW))

### Partition into segments
echo "Partition into segments"
COUNT_LINES=`wc -l $FILE_LIST | awk '{print $1}'`
STAP=$(( COUNT_LINES / FLOW ))
LINE_NUM="1"
for NUM in `seq $FLOW`; do
	tail -n +$LINE_NUM $FILE_LIST | head -n $STAP > ${FILE_LIST_PREFIX}$NUM
	LINE_NUM=$(( LINE_NUM + STAP ))
done
LAST_FLOW=$(( NUM + 1 ))
tail -n +$LINE_NUM $FILE_LIST > ${FILE_LIST_PREFIX}$LAST_FLOW

### upload files
echo "Upload files"
for NUM in `seq $FLOW` $LAST_FLOW; do
	echo "Start flow $NUM"
	( for LOCAL_FILE in `cat ${FILE_LIST_PREFIX}$NUM`; do
		REMOTE_FILE=`echo $LOCAL_FILE | sed -e "s|^$DATA_PATH||" -e 's|^/||'`
		curl -X POST --data-binary @$LOCAL_FILE  $UPLOAD_URL/$REMOTE_FILE > /dev/null 2>&1
	done ) &
done
