#!/usr/bin/env bash
# Strict Mode - http://redsymbol.net/articles/unofficial-bash-strict-mode/
set -euo pipefail
IFS=$'\n\t'

PROD_ACCOUNT_CLUSTERS=(
  eks-pac-prod-eu
  eks-pac-prod-us
  eks-publish-prod-eu
  eks-publish-prod-us
  eks-delivery-prod-eu
  eks-delivery-prod-us

)


SCRIPT_NAME=ops-eks-kubeconfig.sh

#URL to download the kubeconfigs
BUCKET_PROD_URL=https://upp-kubeconfig-ops-469211898354.s3-eu-west-1.amazonaws.com

# Check if the bucket is reachable and get kubeconfigs
STATUS_PROD=$(curl -s --head -w %{http_code} "${BUCKET_PROD_URL}"/check -o /dev/null)


if [ "${STATUS_PROD}" -ne "200" ]
then
  echo "S3 bucket in PROD account not reachable! Check your connection!"
  exit 1
else
TMP_EXEC_DIR="$(mktemp -d /tmp/"${SCRIPT_NAME}".XXXXXXXXXX)" || exit 1
fi

for cluster in "${PROD_ACCOUNT_CLUSTERS[@]}"; do
curl -s "${BUCKET_PROD_URL}"/$cluster > $TMP_EXEC_DIR/$cluster
done


#Clean old kubeconfig
sudo rm -f /etc/skel/content-k8s-auth-setup/eks-kubeconfig

#Merge kubeconfigs
cd "${TMP_EXEC_DIR}"/
KUBECONFIG=$(echo "${PROD_ACCOUNT_CLUSTERS[@]}" | sed 's/ /:/g') kubectl config view --merge=true --flatten=true > eks-kubeconfig
cp eks-kubeconfig "${HOME}"/content-k8s-auth-setup/eks-kubeconfig


#Cleanup
cd - 1>/dev/null
rm -rf "${TMP_EXEC_DIR}"/

# Get kubectx
if [ ! -f /usr/local/bin/kubectx ]; then
sudo sh -c 'curl -s https://raw.githubusercontent.com/ahmetb/kubectx/master/kubectx > /usr/local/bin/kubectx'
sudo chmod 755 /usr/local/bin/kubectx
fi

#echo "New merged kubeconfig generated in $HOME/content-k8s-auth-setup/eks-kubeconfig"
