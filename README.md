# Full node client of smartBCH

This repository contains the code of the full node client of smartBCH, an EVM&amp;Web3 compatible sidechain for Bitcoin Cash.

You can get more information at [smartbch.org](https://smartbch.org).

We are actively developing smartBCH and a testnet will launch soon. Before that, you can [download the source code](https://github.com/smartbch/smartbch/releases/tag/v0.1.0) and start [a private single node testnet](https://docs.smartbch.org/smartbch/deverlopers-guide/runsinglenode) to test your DApp.


### Docker

To run smartBCH via `docker-compose` you can execute the commands below! Note, the first time you run docker-compose it will take a while, as it will need to build the docker image.

```
# Generate a set of 10 test keys.
docker-compose run smartbch gen-test-keys -n 10

# Init the node, include the keys from the last step as a comma separated list.
docker-compose run smartbch init mynode --chain-id 0x1 --init-balance=10000000000000000000 --test-keys="KEY1,KEY2,KEY3,ETC"

# Start it up, you are all set!
docker-compose up
```
