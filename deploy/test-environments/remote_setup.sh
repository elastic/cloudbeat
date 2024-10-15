#!/bin/bash

# Default values
user="ubuntu"

# Parse options
while getopts ":k:s:u:h:d:c:" opt; do
    case $opt in
    k)
        key="$OPTARG"
        ;;
    s)
        src_file="$OPTARG"
        ;;
    u)
        user="$OPTARG"
        ;;
    h)
        host="$OPTARG"
        ;;
    d)
        dest_file="$OPTARG"
        ;;
    c)
        command="$OPTARG"
        ;;
    \?)
        echo "Invalid option -$OPTARG" >&2
        exit 1
        ;;
    esac
done

# Ensure all mandatory parameters are provided
if [ -z "$key" ] || [ -z "$src_file" ] || [ -z "$host" ] || [ -z "$dest_file" ] || [ -z "$command" ]; then
    echo "Usage: $0 -k <key> -s <source_file> -h <host> -d <destination_file> -c <command> [-u <user>]"
    exit 1
fi

# Set the permission for the key file
chmod 600 "$key"

# Copy the file to the EC2/VM instance
scp -o StrictHostKeyChecking=no -v -i "$key" "$src_file" "$user@$host:$dest_file"

# Run the command on the remote EC2/VM instance
ssh -o StrictHostKeyChecking=no -v -i "$key" "$user@$host" "$command"
