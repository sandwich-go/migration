#!/bin/bash

source  ~/.bash_profile
TAG=$1
PROC=migration

IMAGE="zhongtai/${PROC}"

#local
LOCAL_IMAGE="harbor.centurygame.com/${IMAGE}:${TAG}"
export DOCKER_BUILDKIT=1 && docker build -t ${LOCAL_IMAGE} -f ./k8s/server/Dockerfile --ssh=default=$HOME/.ssh/id_rsa .
docker push ${LOCAL_IMAGE}
echo -e "\033[31m local push ${LOCAL_IMAGE} succ \033[0m";


#aws
LATEST_IMAGE="${IMAGE}:latest"
VERSION_IMAGE="${IMAGE}:${TAG}"
AWS_DEFAULT_REGION="us-east-1"
AWS_ECR_URI="public.ecr.aws/c4n2t7d7"
#pip3 install awscli
aws configure set aws_access_key_id "${AWS_ACCESS_KEY_ID}"
aws configure set aws_secret_access_key "${AWS_SECRET_ACCESS_KEY}"
aws --region "${AWS_DEFAULT_REGION}" ecr-public get-login-password | docker login --username AWS --password-stdin "${AWS_ECR_URI}"

export DOCKER_BUILDKIT=1 && docker build -t ${IMAGE} -f ./k8s/server/Dockerfile  --ssh=default=$HOME/.ssh/id_rsa .
docker tag "${LATEST_IMAGE}" "${AWS_ECR_URI}/${LATEST_IMAGE}"
echo "tag ${LATEST_IMAGE}" "${AWS_ECR_URI}/${LATEST_IMAGE}"
docker tag "${LATEST_IMAGE}" "${AWS_ECR_URI}/${VERSION_IMAGE}"
echo "tag ${VERSION_IMAGE}" "${AWS_ECR_URI}/${VERSION_IMAGE}"
docker push "${AWS_ECR_URI}/${LATEST_IMAGE}"
docker push "${AWS_ECR_URI}/${VERSION_IMAGE}"
echo -e "\033[31m aws push ${AWS_ECR_URI}/${VERSION_IMAGE} succ \033[0m";
aws --region "${AWS_DEFAULT_REGION}" ecr-public describe-images --repository-name "zhongtai/${PROC}"

#txCloud


