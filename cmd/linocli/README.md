# Basic run through of using linocli....

To keep things clear, let's have two shells...

`$` is for linocoin (server), `%` is for linocli (client)

## Set up your linocli with a new key

```
% export BCHOME=~/.democli
% linocli keys new coolcool
```

And set up a few more keys for fun...

```
% linocli keys new friend
% linocli keys list
```

## Set up a clean linocoin, initialized with your account

```
$ export BCHOME=~/.demoserve
$ linocoin init
$ linocoin start
```

## Connect your linocli the first time

```
% linocli init --chain-id test_chain_id --node tcp://localhost:46657
```

## Register

```
% linocli tx register --name=friend
```

## Check your balances...

```
% linocli query account --username friend
```

## Send the money

```
% linocli tx send --name demo --amount 1000mycoin --sequence 1 --to $YOU
-> copy hash to HASH
% linocli query tx $HASH
% linocli query account $YOU
```

## Send a post

```
% linocli tx post --name=tuyukai --title=yukaitu --content=tuyukai --postseq=1
```

## Query a post

```
% linocli query post --postauthor=tuyukai --postseq=1
```

## Like a post

```
% linocli tx like --postauthor=tuyukai --postseq=1 --weight=100 --name=tuyukai
```

## Query a like

```
% linocli query like --username=tuyukai --postauthor=tuyukai --postseq=1
```

## Donate a post

```
% linocli tx donate --name=cool --postauthor=tuyukai --amount=100mycoin --sequence=1 --postseq=1
```