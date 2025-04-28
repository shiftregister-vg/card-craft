#!/bin/bash

# Rotate log files
for log in .devbox/logs/*.log; do
  if [ -f "$log" ]; then
    # Create backup with timestamp
    timestamp=$(date +%Y%m%d%H%M%S)
    mv "$log" "${log}.${timestamp}"
    # Create new empty log file
    touch "$log"
  fi
done

echo "Log files rotated successfully" 