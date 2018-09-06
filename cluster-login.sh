#!/bin/bash
output=$(kubectl-login $1 | tee /dev/stderr)

kubectlLoginOutput=($output)

# Output returns the redirect url (for Ops) and the kubeconfig
# Get the kubeconfig path and set the KUBECONFIG environment varible
if [ ${#kubectlLoginOutput[@]} -eq 2  ]
then
    export KUBECONFIG=${output} | awk '{print $2; }'
    echo "Logged in to $1. Using KUBECONFIG=$KUBECONFIG"
# Output returns the kubeconfig path which is set an the KUBECONFIG environment varible
elif [ ${#kubectlLoginOutput[@]} -eq 1 ]
then
    export KUBECONFIG=${output}
    echo "Logged in to $1. Using KUBECONFIG=$KUBECONFIG"
else
    echo "$output"
fi
