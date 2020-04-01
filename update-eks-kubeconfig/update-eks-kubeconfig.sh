#!/usr/bin/env bash
# Strict Mode - http://redsymbol.net/articles/unofficial-bash-strict-mode/
set -euo pipefail
IFS=$'\n\t'

CLUSTERS=(
  pac-staging-eks-eu
)

SCRIPT_NAME=$0

#URL to download the kubeconfigs
BUCKET_URL=https://upp-kubeconfig.s3-eu-west-1.amazonaws.com

# Check if the bucket is reachable and get kubeconfigs
status=$(curl -s --head -w %{http_code} $BUCKET_URL/${CLUSTERS[0]} -o /dev/null)
if [ $status -ne "200" ] 
then
  echo "S3 bucket not reachable! Check your connection!"
  exit 1
else
TMP_EXEC_DIR="$(mktemp -d /tmp/"${SCRIPT_NAME}".XXXXXXXXXX)" || exit 1

for cluster in "${CLUSTERS[@]}"; do
curl -s $BUCKET_URL/$cluster > $TMP_EXEC_DIR/$cluster
done
fi

#Merge kubeconfigs
cd $TMP_EXEC_DIR/
KUBECONFIG=$(echo "${CLUSTERS[@]}" | sed 's/ /:/g') kubectl config view --merge=true --flatten=true > eks-kubeconfig
mv eks-kubeconfig $HOME/.kube/eks-kubeconfig

#Cleanup
cd - 1>/dev/null
rm -rf $TMP_EXEC_DIR/

echo "New merged kubeconfig generated in $HOME/.kube/eks-kubeconfig"
