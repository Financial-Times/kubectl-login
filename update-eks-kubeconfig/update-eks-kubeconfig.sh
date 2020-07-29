#!/usr/bin/env bash
# Strict Mode - http://redsymbol.net/articles/unofficial-bash-strict-mode/
set -euo pipefail
IFS=$'\n\t'

PROD_ACCOUNT_CLUSTERS=(
  eks-pac-test-eu
  eks-pac-test-us
  eks-pac-staging-us
  eks-pac-staging-eu
  eks-pac-prod-us
  eks-pac-prod-eu
  eks-publish-staging-us
  eks-publish-staging-eu
  eks-publish-prod-us
  eks-publish-prod-eu
  eks-delivery-staging-us
  eks-delivery-staging-eu
)
TEST_ACCOUNT_CLUSTERS=(
  eks-delivery-test-eu
  eks-delivery-dev-eu
  eks-publish-test-eu
  eks-publish-dev-eu
)


SCRIPT_NAME=$0

#URL to download the kubeconfigs
BUCKET_PROD_URL=https://upp-kubeconfig-469211898354.s3-eu-west-1.amazonaws.com
BUCKET_DEV_URL=https://upp-kubeconfig-070529446553.s3-eu-west-1.amazonaws.com

# Check if the bucket is reachable and get kubeconfigs
STATUS_PROD=$(curl -s --head -w %{http_code} "${BUCKET_PROD_URL}"/check -o /dev/null)
STATUS_DEV=$(curl -s --head -w %{http_code} "${BUCKET_DEV_URL}"/check -o /dev/null)

if [ "${STATUS_PROD}" -ne "200" ] 
then
  echo "S3 bucket in PROD account not reachable! Check your connection!"
  exit 1
elif [ "${STATUS_DEV}" -ne "200" ]
then
  echo "S3 bucket in DEV account not reachable! Check your connection!"
  exit 1
else
TMP_EXEC_DIR="$(mktemp -d /tmp/"${SCRIPT_NAME}".XXXXXXXXXX)" || exit 1

for cluster in "${PROD_ACCOUNT_CLUSTERS[@]}"; do
curl -s "${BUCKET_PROD_URL}"/$cluster > $TMP_EXEC_DIR/$cluster
done

for cluster in "${TEST_ACCOUNT_CLUSTERS[@]}"; do
curl -s "${BUCKET_DEV_URL}"/$cluster > $TMP_EXEC_DIR/$cluster
done
fi

#Clean old kubeconfig
rm -f "${HOME}"/.kube/eks-kubeconfig

#Merge kubeconfigs
cd "${TMP_EXEC_DIR}"/
KUBECONFIG=$(echo "${PROD_ACCOUNT_CLUSTERS[@]}" "${TEST_ACCOUNT_CLUSTERS[@]}" | sed 's/ /:/g') kubectl config view --merge=true --flatten=true > eks-kubeconfig
mv eks-kubeconfig "${HOME}"/.kube/eks-kubeconfig

#Cleanup
cd - 1>/dev/null
rm -rf "${TMP_EXEC_DIR}"/

echo "New merged kubeconfig generated in "${HOME}"/.kube/eks-kubeconfig"


