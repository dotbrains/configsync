#!/bin/bash
# Development server script for ConfigSync documentation
echo "Starting Jekyll development server..."
echo "Site will be available at: http://localhost:4000/"
echo "Press Ctrl+C to stop the server"
echo ""
bundle exec jekyll serve --incremental --config _config.yml,_config_dev.yml --port 4000
