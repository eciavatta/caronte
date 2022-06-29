#!/bin/bash - 
#===============================================================================
#
#          FILE: feedCaronte.sh
# 
#         USAGE: ./feedCaronte.sh PCAP_DIR_PATH
# 
#   DESCRIPTION: 
# 
#       OPTIONS: ---
#  REQUIREMENTS: inotify-tools, curl
#          BUGS: ---
#         NOTES: test in Debian Buster
#        AUTHOR: Andrea Giovine (AG), 
#  ORGANIZATION: 
#       CREATED: 17/08/2020 16:36:57
#      REVISION:  ---
#===============================================================================

set -o nounset                              # Treat unset variables as an error

CHECK_INOTIFY=$(dpkg-query -W -f='${status}' 'inotify-tools')

if [[ "$CHECK_INOTIFY" != 'install ok installed' ]]; then
	echo "Install inotify-tools"
	exit 1
fi

CHECK_CURL=$(dpkg-query -W -f='${Status}' 'curl')

if [[ "$CHECK_CURL" != 'install ok installed' ]]; then
	echo "Install curl"
	exit 1
fi

if [[ "$#" -ne 1 ]]; then
	echo "Need 1 arg"
	exit 2
fi

PCAP_DIR="$1"

if [[ -z "$PCAP_DIR" ]]; then
	echo "Need path to dir where are store pcaps"
	exit 2
fi

inotifywait -m "$PCAP_DIR" -e close_write -e moved_to |
           while read dir action file; do
             echo "The file $file appeared in directory $dir via $action"
             curl -F "file=@$file" "http://localhost:3333/api/pcap/upload"
           done
