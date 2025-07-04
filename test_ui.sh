#!/bin/bash

# Test script to run the UI in a tmux session
echo "Starting tmux session for UI testing..."

# Create or attach to test session
tmux new-session -d -s ui-test -x 120 -y 40 2>/dev/null || tmux attach-session -t ui-test

# Send commands to the session
tmux send-keys -t ui-test 'cd /home/manuel/workspaces/2025-07-03/extract-help-system-take-2/glazed' Enter
tmux send-keys -t ui-test './glaze help --ui' Enter

# Wait a moment
sleep 1

# Capture the screen to see what's happening
echo "Capturing tmux session screen content:"
tmux capture-pane -t ui-test -p

echo ""
echo "To interact with the UI, run: tmux attach-session -t ui-test"
echo "To kill the session, run: tmux kill-session -t ui-test"
