#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
sevdir="$workspace/src/github.com/severeum"
if [ ! -L "$sevdir/go-severeum" ]; then
    mkdir -p "$sevdir"
    cd "$sevdir"
    ln -s ../../../../../. go-severeum
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$sevdir/go-severeum"
PWD="$sevdir/go-severeum"

# Launch the arguments with the configured environment.
exec "$@"
