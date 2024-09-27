#!/bin/bash
cp /run/secrets/k3s.env /app/.env
/app/server
