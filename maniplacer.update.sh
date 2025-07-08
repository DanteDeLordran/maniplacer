#!/bin/bash
set -e
echo "Waiting for old process to exit..."

while lsof "/home/lordran/workspace/maniplacer/maniplacer" &>/dev/null; do
    sleep 1
done

echo "Replacing old binary..."
mv "/home/lordran/workspace/maniplacer/maniplacer" "/home/lordran/workspace/maniplacer/maniplacer.backup" 2>/dev/null || true
mv "/home/lordran/workspace/maniplacer/maniplacer.update" "/home/lordran/workspace/maniplacer/maniplacer"
chmod +x "/home/lordran/workspace/maniplacer/maniplacer"
rm -f "/home/lordran/workspace/maniplacer/maniplacer.update"

echo "Update complete."

# Optional: Uncomment to auto-restart
# exec "/home/lordran/workspace/maniplacer/maniplacer" "$@"
