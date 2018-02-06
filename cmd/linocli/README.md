# Basic run through of using linocli....

To keep things clear, let's have two shells...

`$` is for linocoin (server), `%` is for linocli (client)

## Set up your linocli with a new key

```
% export BCHOME=~/.democli
% linocli keys new cool
% linocli keys get cool -o json
```

And set up a few more keys for fun...

```
% linocli keys new friend
% linocli keys list
% ME=$(linocli keys get cool | awk '{print $2}')
% YOU=$(linocli keys get friend | awk '{print $2}')
```

## Set up a clean linocoin, initialized with your account

```
$ export BCHOME=~/.demoserve
$ linocoin init $ME
$ linocoin start
```

## Connect your linocli the first time

```
% linocli init --chain-id test_chain_id --node tcp://localhost:46657
```

## Check your balances...

```
% linocli query account $ME
% linocli query account $YOU
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
% linocli tx post --name cool --title title --postSeq 1 --content asdf
```

## Query a post

```
% linocli query post --postAuthor=$ME --postSeq=1
```

## Like a post

```
% linocli tx like --postAuthor=$ME --postSeq=1 weight=100 --name=cool
```

## Query a like

```
% linocli query like --address=$ME --postAuthor=$ME --postSeq=1
```

## Donate a post

```
% linocli tx donate --name=cool --postAuthor=$ME --amount=100mycoin --sequence=1 --postSeq=1
```