#!/bin/bash

#########################################
###                                   ###
### Script for dumping anteater files ###
###                                   ###
#########################################

### Vars
DB_DIR="$1"
BACKUPS_DIR="${2:-/opt/BACKUP/anteater}"
DATE=`date +%Y%m%d-%H%M`
DUMP_DIR="$BACKUPS_DIR/${DATE}_dump"

### Check
if [ -z "$DB_DIR" ] || [ ! -d "$DB_DIR" ]; then
	cat <<EOF
Usage: $0 path/to/anteater/db/dir [path/to/backups/dir]
EOF
	exit 1
fi

### Backup
install -o root -m 0750 -d $BACKUPS_DIR $DUMP_DIR &&
cp -rp $DB_DIR/file.index $DUMP_DIR/ &&
cp -rp $DB_DIR/file.data.* $DUMP_DIR/
