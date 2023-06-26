# Demo

This demonstration will encompass the following topics:
1. Initiating a local Avail + Avail Settlement Layer development network.
2. Establishing a connection between the MetaMask wallet and the Settlement Layer chain.
3. Deploying and minting ERC20 tokens and NFT tokens.
4. Importing assets into the MetaMask wallet.

## Setting up a Linux Environment in Docker (Optional)

The examples in this guide are for linux/arm64.

If you are using a different operating system like macOS or Windows, you can set up a Linux environment in a Docker container.
Here are the steps to follow:

1. Run the following commands to prepare a Docker container:
```shell
docker run -d -it -p 22:22 ubuntu
docker exec -it <container-id> bash
```
2. Inside the Docker container, run the following commands to install necessary dependencies:
```shell
apt update && apt install openssh-server sudo unzip curl jq git vim -y
useradd -rm -d /home/ubuntu -s /bin/bash -g root -G sudo -u 1000 ubuntu
echo 'ubuntu:ubuntu' | chpasswd
service ssh start
```
3. Install Node.js:
```shell
curl -sL https://deb.nodesource.com/setup_18.x -o /tmp/nodesource_setup.sh
sudo bash /tmp/nodesource_setup.sh
apt install nodejs
```

## Setting up Avail

1. Download the Avail release from the [Releases page](https://github.com/availproject/avail/releases)
```shell
curl -LO https://github.com/availproject/avail/releases/download/v1.3.0-rc3/data-avail-linux-aarch64.tar.gz
```

2. Extract the downloaded tar.gz file:
```shell
tar -xzvf data-avail-linux-aarch64.tar.gz
```
3. Start the Avail development network:
```shell
./data-avail-linux-aarch64 --dev
```

## Setting up a Local SL DevNet

To set up a local DevNet, follow these steps:
1. Generate an auth token from [github.com/settings/tokens](github.com/settings/tokens).
2. Save the generated token as the G_TOKEN environment variable.
3. Run the following commands to download the Avail Settlement Layer binary, unzip it, and start a DevNet:
```shell
ASSET_ID=$(curl -H "Authorization: token $G_TOKEN" https://api.github.com/repos/availproject/op-evm/releases/tags/v0.0.1 | jq '.assets[] | select(.name == "op-evm-linux-arm64.zip") | .id')
curl -LJO -H "Authorization: token $G_TOKEN" -H 'Accept: application/octet-stream' https://api.github.com/repos/availproject/op-evm/releases/assets/$ASSET_ID
unzip op-evm-linux-arm64.zip
mkdir -p data/test-accounts
./op-evm devnet
```

You should see logs similar to this:
```
2023-06-16T12:13:09.921+0300 [INFO]  all nodes started: servers_count=3
2023-06-16T12:13:09.921+0300 [INFO]  polygon.server.avail: About to process node staking...: node_type=watchtower
| NODE TYPE           | JSONRPC URL             | FRAUD SERVER URL        | GRPC ADDR       |
| bootstrap-sequencer | http://127.0.0.1:49601/ | http://127.0.0.1:49604/ | 127.0.0.1:49602 |
| sequencer           | http://127.0.0.1:49605/ | http://127.0.0.1:49608/ | 127.0.0.1:49606 |
| watchtower          | http://127.0.0.1:49609/ | http://127.0.0.1:49612/ | 127.0.0.1:49610 |
```
Copy the first json url from the table. (in our case it's `http://127.0.0.1:49601/`).

## Set up an AWS network

To deploy a devnet or a testnet in AWS using terraform follow the instructions [here](../deployment/readme.md).

## Connecting MetaMask to the Network

1. Open your MetaMask wallet.
2. Go to **Settings > Networks > Add Network > Add a network manually**.
3. Fill in the following network details:
    - Network name: `Avail SL`
    - New RPC URL: `http://127.0.0.1:49601/` (use your bootstrap sequencer rpc link)
    - Chain ID: `100`
    - Currency symbol: `ETH`
4. Click "Save" and switch to the `Avail Sl` network.

## Transferring Tokens from the Faucet Account

To transfer funds from the faucet account to your account, you need to import the faucet account in your MetaMask wallet and send tokens. Here are the steps:
1. Copy the provided private key: `e29fc399e151b829ca68ba811108965aeec52c21f2ac1744cb28f203231dc085`
2. Open your MetaMask wallet.
3. Click your account icon and press **Import account**.
4. Paste the private key in the field and click "Import".
5. You should see that you have tokens in the current account.
6. To send tokens to your initial account, click the **Send** button, select **Transfer between my accounts**, choose **Account 1**, enter the amount, and confirm the transaction.

You should see your transaction in the queue, after it is processed switch to **Account 1**

## Exporting Private Key

To export the private key from your MetaMask wallet, follow these steps:
1. In your MetaMask wallet, click the **Three dots** next to your account name and key.
2. Go to **Account details > Export private key**.
3. Enter your password and confirm.
4. Copy your private key and click "Done".

## Setting up Hardhat

1. Clone the Avail settlement contracts repository:
```shell
git clone https://$G_TOKEN@github.com/availproject/op-evm-contracts.git
cd op-evm-contracts/testing
```
2. Install dependencies and copy the environment file:
```shell
cd op-evm-contracts/testing
npm install
cp .env.example .env
```
3. Set the appropriate values for the environment variables in the .env file. In this demo, the values should be as follows:
```shell
OPEVM_URL=http://127.0.0.1:49601/
ACC_PRIVATE_KEY=<private key>
```

## Deploying and Minting NFT

To deploy and mint an NFT, run the following command:
```shell
npx hardhat run scripts/deploy-nft.ts
```
You should see a response similar to this:
```
TestNFT deployed to 0x8b4cdD710bB9A4f02578c349DB5d1DF7FeFe9343; token uri is: ipfs://bafybeig37ioir76s7mg5oobetncojcm3c3hxasyd4rvid4jqhy4gkaheg4/?filename=0-PUG.json; tokenCounter is: 0
```

### Importing NFTs in MetaMask

To import NFTs into MetaMask, follow these steps:
1. Go to MetaMask and navigate to the NFTs tab.
2. Click **Import NFT**.
3. Fill in the fields with the provided information.
    - Address: `0x8b4cdD710bB9A4f02578c349DB5d1DF7FeFe9343`
    - Token ID: `0`
4. Click **Add**.

## Deploying and Minting ERC20 Token

To deploy and mint an ERC20 token, run the following command:
```shell
npx hardhat run scripts/deploy-token.ts
```

You should see a response similar to this:
```
TestToken deployed to 0x282a5B9D1bcD5Ef7445e94bb499A3F9CDB06eAE8
```

### Importing ERC20 Tokens in MetaMask

To import ERC20 tokens into your MetaMask wallet, follow these steps:
1. Go to your MetaMask wallet and navigate to the Assets tab.
2. Click **Import tokens**.
3. Fill in the fields with the provided information.
    - Token contract address: `0x256828c8E6CD0AC05FB1f31aFFaf17248fD863F9`
    - Token symbol: `TEST`
    - Token decimal: `18`
4. Click **Add custom token** and **Import tokens**.
