#!/bin/bash
# Helper script to check OpenCode connection status

echo "=== OpenCode Connection Diagnostics ==="
echo ""

# Check if OpenCode is running
echo "1. Checking if OpenCode is running..."
if ps aux | grep -v grep | grep "opencode" > /dev/null; then
    echo "   ✓ OpenCode process found"
else
    echo "   ✗ OpenCode not running"
fi
echo ""

# Check if OpenCode is listening on a port
echo "2. Checking if OpenCode is listening on a port..."
OPENCODE_PORTS=$(lsof -i -P | grep opencode | grep LISTEN)
if [ -z "$OPENCODE_PORTS" ]; then
    echo "   ✗ No OpenCode server listening"
    echo "   → Make sure OpenCode was started with --port flag"
    echo "   → Command: opencode --port 0 ."
else
    echo "   ✓ OpenCode server listening:"
    echo "$OPENCODE_PORTS" | awk '{print "     Port:", $9}'
fi
echo ""

# Check if in tmux
echo "3. Checking tmux environment..."
if [ -n "$TMUX" ]; then
    echo "   ✓ Running inside tmux"
    
    # Check for opencode window
    if tmux list-windows -F "#{window_name}" | grep -q "opencode"; then
        echo "   ✓ Found 'opencode' tmux window"
        
        # Show what's running in that window
        WINDOW_CMD=$(tmux list-windows -F "#{window_name}:#{pane_current_command}" | grep opencode)
        echo "   → Window: $WINDOW_CMD"
    else
        echo "   ✗ No 'opencode' tmux window found"
    fi
else
    echo "   ✗ Not running in tmux"
fi
echo ""

# Check OpenCode session file
echo "4. Checking for OpenCode session info..."
OPENCODE_DIR="$HOME/.local/share/opencode"
if [ -d "$OPENCODE_DIR" ]; then
    echo "   ✓ OpenCode data directory exists"
    
    # Find recent session files
    RECENT_SESSIONS=$(find "$OPENCODE_DIR" -name "*.json" -mtime -1 2>/dev/null | head -5)
    if [ -n "$RECENT_SESSIONS" ]; then
        echo "   ✓ Found recent session files"
    fi
else
    echo "   ✗ OpenCode data directory not found"
fi
echo ""

# Summary
echo "=== Summary ==="
if [ -n "$OPENCODE_PORTS" ] && [ -n "$TMUX" ]; then
    echo "✓ OpenCode should be connectable from neovim"
    echo ""
    echo "Next steps:"
    echo "1. In neovim, try: <leader>oa (Ask OpenCode)"
    echo "2. Should automatically switch to OpenCode window"
else
    echo "✗ OpenCode is not properly configured"
    echo ""
    echo "To fix:"
    if [ -z "$OPENCODE_PORTS" ]; then
        echo "1. Kill current OpenCode: killall opencode"
        echo "2. Start with port: opencode --port 0 ."
    fi
    if [ -z "$TMUX" ]; then
        echo "1. Use 'sesh' to start your project session"
    fi
fi
