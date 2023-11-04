#!/bin/bash
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 151654911502.dkr.ecr.us-west-2.amazonaws.com
docker build -t bluelytics_api .
docker tag bluelytics_api:latest 151654911502.dkr.ecr.us-west-2.amazonaws.com/bluelytics_api:latest
docker push 151654911502.dkr.ecr.us-west-2.amazonaws.com/bluelytics_api:latest
