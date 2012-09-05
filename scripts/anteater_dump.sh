#!/bin/bash

#########################################
###                                   ###
### Script for dumping anteater files ###
###                                   ###
#########################################

### Vars
BACKUPS_DIR="${1:-/opt/BACKUP/anteater}"
USER="${2:-at}"
GROUP="$USER"
DATE=`date +%Y%m%d-%H%M`
DUMP_DIR="$BACKUPS_DIR/${DATE}_dump"

### Check
case $1 in
	help|h|-h|--help)
		cat <<EOF
Usage: $0 [path/to/backups/dir] [anteater_user]
EOF
		exit 1
		;;
esac

### Backup
echo "install -o $USER -g $GROUP -m 0750 -d $BACKUPS_DIR $DUMP_DIR &&"
echo "aecommand dump $DUMP_DIR"
