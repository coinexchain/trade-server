- 查询区块时间

```
curl -X GET "http://localhost:8000/misc/block-times?height=500&count=2"  -H "accept: application/json" 
[1566283591,1566283589]%
```

- 查询用户签名的tx / 查询用户的income交易

```bash
curl -X GET "http://localhost:8000/tx/txs?account=coinex15tt4lj3fa93f54mzdl2aamfjv0m42hv8dr0x38&time=1566282669&count=2&sid=0"  -H "accept: application/json"
curl -X GET "http://localhost:8000/tx/incomes?account=coinex15tt4lj3fa93f54mzdl2aamfjv0m42hv8dr0x38&time=1566282669&count=2&sid=0"  -H "accept: application/json"
{
  "data": [
    {
      "signers": [
        "coinex15tt4lj3fa93f54mzdl2aamfjv0m42hv8dr0x38"
      ],
      "transfers": [
        {
          "sender": "coinex15tt4lj3fa93f54mzdl2aamfjv0m42hv8dr0x38",
          "recipient": "coinex17xpfvakm2amg962yls6f84z3kell8c5lm7j9tl",
          "amount": "1000000000000cet"
        },
        {
          "sender": "coinex16kfcdc9wgd0zjta7p67dh92twhk4lvujls8lmk",
          "recipient": "coinex15tt4lj3fa93f54mzdl2aamfjv0m42hv8dr0x38",
          "amount": "57896044618658097711785492504343953926634992332820282019728792003956564819967abc"
        }
      ],
      "serial_number": 1,
      "msg_types": [
        "MsgIssueToken"
      ],
      "tx_json": "{\"msg\":[{\"name\":\"ABC Token\",\"symbol\":\"www\",\"total_supply\":\"57896044618658097711785492504343953926634992332820282019728792003956564819967\",\"owner\":\"coinex15tt4lj3fa93f54mzdl2aamfjv0m42hv8dr0x38\",\"mintable\":true,\"burnable\":true,\"addr_forbiddable\":false,\"token_forbiddable\":false,\"url\":\"www.abc.org\",\"description\":\"token abc is a example token\",\"identity\":\"552A83BA62F9B1F8\"}],\"fee\":{\"amount\":[{\"denom\":\"cet\",\"amount\":\"100000000\"}],\"gas\":200000},\"signatures\":[{\"pub_key\":[3,43,37,122,69,38,236,108,104,9,202,92,199,201,181,119,189,162,2,117,251,8,163,125,190,33,2,167,231,212,163,14,132],\"signature\":\"FLfeWm27tf7rtMHjVa228UwAA03fiPplFxsuGxq0IhkolKENVQg1dMtdlsvybfgIfnL2E1ArSsyNoDlTnSHB7w==\"}],\"memo\":\"\"}",
      "height": 57
    },
    {
      "signers": [
        "coinex15tt4lj3fa93f54mzdl2aamfjv0m42hv8dr0x38"
      ],
      "transfers": [
        {
          "sender": "coinex15tt4lj3fa93f54mzdl2aamfjv0m42hv8dr0x38",
          "recipient": "coinex17xpfvakm2amg962yls6f84z3kell8c5lm7j9tl",
          "amount": "1000000000000cet"
        },
        {
          "sender": "coinex16kfcdc9wgd0zjta7p67dh92twhk4lvujls8lmk",
          "recipient": "coinex15tt4lj3fa93f54mzdl2aamfjv0m42hv8dr0x38",
          "amount": "57896044618658097711785492504343953926634992332820282019728792003956564819967abc"
        }
      ],
      "serial_number": 0,
      "msg_types": [
        "MsgIssueToken"
      ],
      "tx_json": "{\"msg\":[{\"name\":\"ABC Token\",\"symbol\":\"abc\",\"total_supply\":\"57896044618658097711785492504343953926634992332820282019728792003956564819967\",\"owner\":\"coinex15tt4lj3fa93f54mzdl2aamfjv0m42hv8dr0x38\",\"mintable\":true,\"burnable\":true,\"addr_forbiddable\":false,\"token_forbiddable\":false,\"url\":\"www.abc.org\",\"description\":\"token abc is a example token\",\"identity\":\"552A83BA62F9B1F8\"}],\"fee\":{\"amount\":[{\"denom\":\"cet\",\"amount\":\"100000000\"}],\"gas\":200000},\"signatures\":[{\"pub_key\":[3,43,37,122,69,38,236,108,104,9,202,92,199,201,181,119,189,162,2,117,251,8,163,125,190,33,2,167,231,212,163,14,132],\"signature\":\"fxd69H59RvPFFme8C7nYD671D2DAMap81fR7HEae5QZBnxpVPVCQnb6PB8yPTzugtjb5phWEtZqBPi0350y/aA==\"}],\"memo\":\"\"}",
      "height": 10
    }
  ],
  "timesid": [
    1566282645,
    5,
    1566282544,
    1
  ]
}
```

- 查询给定market的tickers
```bash
curl -X GET "http://localhost:8000/market/tickers?market_list=qqq/cet" -H "accept: application/json"
[{"market":"qqq/cet","new":"4.899999998738545890","old":"5.299999998200478143"}]% 
```
- 查询给定market的深度
```bash
curl -X GET "http://localhost:8000/market/depths?market=abc/cet&count=1"
{“sell”:[{“p”:“0.015200000000000000”,“a”:“10000000”}],“buy”:[{“p”:“0.005200000000000000",“a”:“10000000"}]}
```
- 查询给定market的K线信息
```bash
http://localhost:8000/market/candle-sticks?market=abc/cet&timespan=16&time=1566289684&count=1&sid=0
```
- 查询用户的orders信息
```bash
curl -X GET "http://localhost:8000/market/user-orders?account=coinex18t4sp5kv07czmv2ar9ds04z3xx946kkkcwnfpx&time=1566371118&count=3&sid=0" -H "accept: application/json"
{
  "create_order_info": {
    "data": [
      {
        "order_id": "coinex18t4sp5kv07czmv2ar9ds04z3xx946kkkcwnfpx-5",
        "sender": "coinex18t4sp5kv07czmv2ar9ds04z3xx946kkkcwnfpx",
        "trading_pair": "abc/cet",
        "order_type": 2,
        "price": "1.000000000000000000",
        "quantity": 300000000,
        "side": 1,
        "time_in_force": 3,
        "feature_fee": 0,
        "height": 10,
        "frozen_fee": 1000000,
        "freeze": 300000000
      }
    ],
    "timesid": [
      1566371090,
      6
    ]
  },
  "fill_order_info": {
    "data": [
      {
        "order_id": "coinex18t4sp5kv07czmv2ar9ds04z3xx946kkkcwnfpx-5",
        "trading_pair": "abc/cet",
        "height": 12,
        "side": 1,
        "price": "1.000000000000000000",
        "left_stock": 0,
        "freeze": 30000000,
        "deal_stock": 300000000,
        "deal_money": 270000000,
        "curr_stock": 300000000,
        "curr_money": 270000000
      }
    ],
    "timesid": [
      1566371090,
      8
    ]
  },
  "cancel_order_info": {
    "data": [
      {
        "order_id": "coinex18t4sp5kv07czmv2ar9ds04z3xx946kkkcwnfpx-5",
        "trading_pair": "abc/cet",
        "height": 12,
        "side": 1,
        "price": "1.000000000000000000",
        "del_reason": "The order was fully filled",
        "used_commission": 1000000,
        "left_stock": 0,
        "remain_amount": 30000000,
        "deal_stock": 300000000,
        "deal_money": 270000000
      }
    ],
    "timesid": [
      1566371090,
      12
    ]
  }
}
```

- 查询market-deal

```bash
curl -X GET "http://localhost:8000/market/deals?market=abc/cet&time=1566288724&count=2&sid=0" -H "accept: application/json"
{
  "data": [
    {
      "order_id": "coinex15tt4lj3fa93f54mzdl2aamfjv0m42hv8dr0x38-2415",
      "trading_pair": "abc/cet",
      "height": 2836,
      "side": 1,
      "price": "4.900000000000000000",
      "left_stock": 0,
      "freeze": 2,
      "deal_stock": 566449077,
      "deal_money": 2775600476,
      "curr_stock": 566449077,
      "curr_money": 2775600476
    },
    {
      "order_id": "coinex15tt4lj3fa93f54mzdl2aamfjv0m42hv8dr0x38-7164",
      "trading_pair": "abc/cet",
      "height": 2836,
      "side": 2,
      "price": "3.200000000000000000",
      "left_stock": 0,
      "freeze": 0,
      "deal_stock": 505528118,
      "deal_money": 2477087777,
      "curr_stock": 505528118,
      "curr_money": 2477087777
    }
  ],
  "timesid": [
    1566288601,
    33194,
    1566288601,
    33192
  ]
}
```

- 查询bancorlite-trade

```bash
curl -X GET "http://localhost:8000/bancorlite/trades?account=coinex1za0s76cwrkn5lgzmdkznkqplx28avrsvv0qx0q&time=1566269120&count=2&sid=0"  -H "accept: application/json"
{
  "data": [
    {
      "sender": "coinex1za0s76cwrkn5lgzmdkznkqplx28avrsvv0qx0q",
      "stock": "qxq",
      "money": "cet",
      "amount": 100,
      "side": 1,
      "money_limit": 120,
      "transaction_price": "0.000000000000000000",
      "block_height": 9828
    },
    {
      "sender": "coinex1za0s76cwrkn5lgzmdkznkqplx28avrsvv0qx0q",
      "stock": "qxq",
      "money": "cet",
      "amount": 100,
      "side": 1,
      "money_limit": 120,
      "transaction_price": "0.000000000000000000",
      "block_height": 9673
    }
  ],
  "timesid": [
    1566269113,
    25,
    1566268781,
    16
  ]
}
```

- 查询bancorlite-info

```bash
curl -X GET "http://localhost:8000/bancorlite/infos?market=cet/qxq&time=1566271305&count=1&sid=0"  -H "accept: application/json"
{
  "data": [
    {
      "sender": "coinex1cr6ven9f6rqu2ekageepakgpkyxs9rlfe6tmad",
      "stock": "qxq",
      "money": "cet",
      "init_price": "1.000000000000000000",
      "max_supply": "10000000000000",
      "max_price": "5.000000000000000000",
      "price": "1.000000000200000000",
      "stock_in_pool": "9999999999500",
      "money_in_pool": "500",
      "earliest_cancel_time": 1563954165,
      "block_height": 0
    }
  ],
  "timesid": [
    1566269113,
    26
  ]
}
```

- 查询token的股吧信息

```bash
curl -X GET "http://localhost:8000/comment/comments?token=cet&time=1566268004&count=1&sid=0"  -H "accept: application/json"
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   208  100   208    0     0  33052      0 --:--:-- --:--:-- --:--:-- 34666
{
  "data": [
    {
      "id": 0,
      "height": 9268,
      "sender": "coinex1cr6ven9f6rqu2ekageepakgpkyxs9rlfe6tmad",
      "token": "cet",
      "donation": 2,
      "title": "I love cet.",
      "content": "CET to da moon!!!",
      "content_type": 3,
      "references": null
    }
  ],
  "timesid": [
    1566267908,
    5
  ]
}
```

- 查询delegator的unbonding信息

```bash
curl -X GET "http://localhost:8000/expiry/unbondings?account=coinex1te36zqyrpygz384nktprzgtctp2hdq3mx0tn7v&time=1597927914&sid=0&count=1" -H "accept: application/json"
{
  "data": [
    {
      "delegator": "coinex1te36zqyrpygz384nktprzgtctp2hdq3mx0tn7v",
      "validator": "coinexvaloper15tt4lj3fa93f54mzdl2aamfjv0m42hv8kvvwln",
      "amount": "5000",
      "completion_time": "2019-09-10T13:01:18Z"
    }
  ],
  "timesid": [
    1568120478,
    0
  ]
}
```

- 查询用户的unlocks信息

```bash
curl -X GET "http://localhost:8000/expiry/unlocks?account=coinex1fxdvpzaw2qcc7dcxancrvfej3mxvuylhya20x0&time=1566307026&sid=0&count=1"  -H "accept: application/json"
{
  "data": [
    {
      "address": "coinex1fxdvpzaw2qcc7dcxancrvfej3mxvuylhya20x0",
      "unlocked": [
        {
          "denom": "cet",
          "amount": "1000"
        }
      ],
      "locked_coins": null,
      "frozen_coins": [
        {
          "denom": "cet",
          "amount": "3493087654583"
        },
        {
          "denom": "qqq",
          "amount": "1477629660988"
        }
      ],
      "coins": [
        {
          "denom": "cet",
          "amount": "24962026053402470"
        },
        {
          "denom": "qqq",
          "amount": "24998549621705490"
        }
      ],
      "height": 10956
    }
  ],
  "timesid": [
    1566306954,
    8
  ]
}
```

- 查询用户的redelegations信息

```bash
curl -X GET "http://localhost:8000/expiry/redelegations?account=coinex1hclkyahl2fmf0qng4ksz9su6zl4d8q3s4h5t7k&time=1597927914&sid=0&count=1" -H "accept: application/json"
{
  "data": [
    {
      "delegator": "coinex1hclkyahl2fmf0qng4ksz9su6zl4d8q3s4h5t7k",
      "src": "coinexvaloper1k4vfyednnrst48we29phjdu09tprdqv5h8cvga",
      "dst": "coinexvaloper1mxewdqph6qq7vudhjlhq5vcpvp28utnnnedtr2",
      "amount": "300000000000",
      "completion_time": "2019-08-21T06:23:00Z"
    }
  ],
  "timesid": [
    1566368580,
    0
  ]
}
```









