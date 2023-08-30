# filsigner-relayed

## Installation
```
go install github.com/data-preservation-programs/filsigner-relayed/cmd/filsigner@latest
```

## Usage
### Run as native binary
```shell
$ ./filsigner run -h
NAME:
   filsigner run - Run the filsigner server to sign deal proposals

USAGE:
   filsigner run [command options] [arguments...]

OPTIONS:
   --allowed-requester value, -r value [ --allowed-requester value, -r value ]  The peer ID of the allowed requester [$ALLOWED_REQUESTERS]
   --identity-key value, -k value                                               The base64 encoded private key of the peer to use as the identity [$IDENTITY_KEY]
   --sign-key value, -s value [ --sign-key value, -s value ]                    The private key of the address to sign with [$SIGN_KEYS]
   --relay-info value [ --relay-info value ]                                    [Local testing only] The relay info to use to connect to the allowed requesters - this will override the default relay servers from SPADE [$RELAY_INFOS]
   --help, -h                                                                   show help
```
### Run as docker container
```shell
$ docker pull datapreservationprogram/filsigner-relayed:latest
$ ALLOWED_REQUESTERS=<SPADE_PEER>\
    IDENTITY_KEY=<IDENTITY_PRIVATE_KEY>\
    SIGN_KEYS=<WALLET_PRIVATE_KEY_1>,<AND_MORE>\
    docker run -e ALLOWED_REQUESTERS -e IDENTITY_KEY -e SIGN_KEYS datapreservationprogram/filsigner-relayed:latest
```

## Local testing
Below should be put into unit tests, but for now, here's how to test locally.

### Create Relay Daemon
https://github.com/libp2p/go-libp2p-relay-daemon/

Example relay info
```
/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWM9uR8o7yo1GFzGser2iYYzbphJ24qZFLFM6evqXBEvoK
```

### Create test server peer
```
peer id:      12D3KooWD1GwcSwpN9ErEm3TETTUbzveXaMMy1TNEC7XR6vvFQEZ
public key:   CAESIC9eikbtXrtfLuCaoslNkZzb4ak7XYNz5HM9BldoKOcu
private key:  CAESQKh6mFqQytA2KPTQTs7nHf+AI5b8qYT7slPc0OLqU6etL16KRu1eu18u4JqiyU2RnNvhqTtdg3Pkcz0GV2go5y4=
```

### Create test client peer
```
peer id:      12D3KooWS7rfPuvgSx3tXZb5u7oHfzYvv88mtw5caDtpcffgfbnH
public key:   CAESIPI2FTSJ+5tAx1tPXFTlD/kRxlCPsipBcZ1/eFjrqKAq
private key:  CAESQKmKIi9Q38aGg5SKhRlbLLE3l/1lDIk52T5dE2p4K0ed8jYVNIn7m0DHW09cVOUP+RHGUI+yKkFxnX94WOuooCo=
```

### Run server against local relay daemon
```
./filsigner run -r 12D3KooWS7rfPuvgSx3tXZb5u7oHfzYvv88mtw5caDtpcffgfbnH -k CAESQKh6mFqQytA2KPTQTs7nHf+AI5b8qYT7slPc0OLqU6etL16KRu1eu18u4JqiyU2RnNvhqTtdg3Pkcz0GV2go5y4= --relay-info /ip4/127.0.0.1/tcp/4001/p2p/12D3KooWM9uR8o7yo1GFzGser2iYYzbphJ24qZFLFM6evqXBEvoK -s 7b2254797065223a22736563703235366b31222c22507269766174654b6579223a2244485a65316e7146756c7142382b44345a6167566f4f6654566d366e6f45415076414431705051446167343d227d
```

### Run client - success
```
./filsigner test -k CAESQKmKIi9Q38aGg5SKhRlbLLE3l/1lDIk52T5dE2p4K0ed8jYVNIn7m0DHW09cVOUP+RHGUI+yKkFxnX94WOuooCo= -d 12D3KooWD1GwcSwpN9ErEm3TETTUbzveXaMMy1TNEC7XR6vvFQEZ -c f1cbqqzvzx6suldlmxbc33uqjvhkwyjsyvudh3xwi --relay-info /ip4/127.0.0.1/tcp/4001/p2p/12D3KooWM9uR8o7yo1GFzGser2iYYzbphJ24qZFLFM6evqXBEvoK
```

### WalletKeyNotFound
```
./filsigner test -k CAESQKmKIi9Q38aGg5SKhRlbLLE3l/1lDIk52T5dE2p4K0ed8jYVNIn7m0DHW09cVOUP+RHGUI+yKkFxnX94WOuooCo= -d 12D3KooWD1GwcSwpN9ErEm3TETTUbzveXaMMy1TNEC7XR6vvFQEZ -c f1ws3n5tuxtyg26lraqkjirz7qon7y7ckju7hhmii --relay-info /ip4/127.0.0.1/tcp/4001/p2p/12D3KooWM9uR8o7yo1GFzGser2iYYzbphJ24qZFLFM6evqXBEvoK
```

### UnauthorizedRequester
```
./filsigner test -k CAESQPdiErDHCIvc5suvj5+h+iv4vZWcDhLP7wxZL+jlPYiOOrLT5kJ5sDMT4+9jtW6i+oa+FRaaozBGgQG2nHC3dHg= -d 12D3KooWD1GwcSwpN9ErEm3TETTUbzveXaMMy1TNEC7XR6vvFQEZ -c f1cbqqzvzx6suldlmxbc33uqjvhkwyjsyvudh3xwi --relay-info /ip4/127.0.0.1/tcp/4001/p2p/12D3KooWM9uR8o7yo1GFzGser2iYYzbphJ24qZFLFM6evqXBEvoK
```
