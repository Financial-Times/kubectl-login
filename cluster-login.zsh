#/bin/zsh

source "$HOME/.zshrc"
export KUBECONFIG=$(kubectl-login "$1" | tail -1)
echo "Using KUBECONFIG=$KUBECONFIG"
kubectl cluster-info

