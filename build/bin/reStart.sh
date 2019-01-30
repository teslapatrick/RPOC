cd chain/geth
rm -rf `ls |grep -v nodekey`
cd ../..
./geth --datadir=chain/ init chain.json
./geth --datadir=chain --networkid=55555 --rpc --rpcapi="web3, eth, admin, personal, txpool, miner, clique" --rpccorsdomain='*' console
