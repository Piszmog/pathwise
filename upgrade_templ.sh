#!/bin/sh

if [ "$#" -ne 2 ]; then
	echo "Usage: $0 <old version> <new version>"
	exit 1
fi

old_version="$1"
new_version="$2"

LC_CTYPE=C find . \( -path "./dist" -o -path "./.git" \) -prune -o -type f ! -name ".DS_Store" -exec sed -i '' -e "s/$old_version/$new_version/g" {} \;
