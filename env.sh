export COREUM_CHAIN_ID="coreum-testnet-1"
export COREUM_DENOM="utestcore"
export COREUM_NODE="https://full-node.testnet-1.coreum.dev:26657"
export COREUM_VERSION="v1.0.0"

export COREUM_CHAIN_ID_ARGS="--chain-id=$COREUM_CHAIN_ID"
export COREUM_NODE_ARGS="--node=$COREUM_NODE $COREUM_CHAIN_ID_ARGS"

export COREUM_HOME=$HOME/.core/"$COREUM_CHAIN_ID"

#Setup core
export COREUM_PATH=$HOME/.core/crust
mkdir $COREUM_PATH
cd $COREUM_PATH
export PATH="$COREUM_PATH/crust/bin:$PATH"
git clone https://github.com/CoreumFoundation/crust
export PATH="$COREUM_PATH/crust/bin:$PATH"
$COREUM_PATH/crust/bin/crust build images

export COREUM_BINARY_NAME=$(arch | sed s/aarch64/cored-linux-arm64/ | sed s/x86_64/cored-linux-amd64/)