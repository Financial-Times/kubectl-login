#!/bin/bash
output=$(kubectl-login $1 | tee /dev/stderr)

kubectlLoginOutput=($output)

# Output returns the redirect url (for Ops) and the kubeconfig
# Get the kubeconfig path and set the KUBECONFIG environment varible
if [ ${#kubectlLoginOutput[@]} -eq 2  ]
then
    kubeconfigPath=$(echo ${output} | awk '{print $2; }')
    export KUBECONFIG=${kubeconfigPath}
    echo "Logged in to $1. Using KUBECONFIG=$KUBECONFIG"
# Output returns the kubeconfig path which is set an the KUBECONFIG environment varible
elif [ ${#kubectlLoginOutput[@]} -eq 1 ]
then
    export KUBECONFIG=${output}
    echo "Logged in to $1. Using KUBECONFIG=$KUBECONFIG"
fi
