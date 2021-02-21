# PacketSockDgramDualConn

## 目的    

linux上でudpのconnを提供します。 
udpを通常扱う場合は、ipヘッダーは気にしないと思いますが、このconnではipヘッダーを取り扱います。  
それがこの実装の目的です。  
あとipv4/ipv6デュアルスタックをシンプルに実現するために用意しました。  
実現方法は、AF_PACKET、SOCK_DGRAM、ETH_P_IP、ETH_P_IPV6、を指定することで実現しています。
ネットワークインターフェイスを指定せず、ipv4、ipv6の両方からのudpをひとつのconnで受け取れるように実装しています。  

## ちなみに
AF_PACKET、SOCK_DGRAM、ETH_P_ALLの指定だけでひとつのソケットで、対応できることはわかったのですが、udpをふたつ受けとってしまうため、このような実装にしました。  
パッケージ名のDualConnは、仕方なく、ふたつのconnを使ってますよ、というニュアンスのつもりです。  
正確にはconnではないですが...  
ソースコードはソケットつかってrecvしているだけです。素朴な実装となっています。    

## 疑問  
- golangの標準のnetで実現できるのでは？  
はい、その疑念はあります。RawConn使えば、よいのかなといろいろ試したのですが、うまくできませんでした。何か見落としているかもしれません。わかったら教えてください。  

- 性能でるの？  
わかりません。他の実装と比べきれていないので、正直わかりません、実装はひろえたデータグラムをパースして自分宛てのポートかどうかの判定を行っているため、そういった処理が、性能に影響するのだろうか、あとは、他の要因があって実は性能がよくないという疑念はあります。

## 取扱方について
cmd/server.go の実装をみるとわかると思います。処理をブロックするrecvを待ち続けるという方法を利用者にしてもらわないようにしています。ハンドラー関数を用意してもらいそれを設定してもらう形にしています。  
conn生成のために用意している関数は、NewDualConn、NewDualConnIf、NewConnです。受信待ちのReadfromとNewConnは直接利用してもらう想定ではないですが、用意しています。
※関数とメソッドの用語の使い分けをして説明していないことに注意してください。

## 内部では
- goroutine(ゴルーティン)は使っているの？  
はい使っています。合計４つ使っています。syscall.Recvfromでブロックしている箇所にたいして使っています。それとそのgoroutineの中で、contextでのクローズ検知のために使っています。

- ErrNotDestPortが発生するのはなぜ？  
AF_PACKETを使っており、udpの自分宛先のポートを自前でチェックしているからです。エラーをださない案も考えたのですが、ここは利用者ではじいてもらう設計にしています。

- ErrSomethingIsWrongが発生するのはなぜ？  
これは理由がわかっていないのです。受信したパケットの先頭1バイト目でipv4,ipv6を判定しているのですが、判定にもれたものをこのエラーにするようにしています。ETH_P_IP、ETH_P_IPV6を指定しているのでipv4、ipv6のもの以外は受信しないと考えているのですが、届くものがあり、このような仕様にしました。  

- ErrBufsizeTooShortはどういったときに発生するの？
ハンドラ関数を設定する際に受信するさいに使用するバッファのサイズが受信したパケットより小さい場合に発生します。この実装は、ipヘッダー、udpヘッダーのサイズも含むためのそのサイズも考慮したサイズを設定する必要があります。udpどれぐらいのサイズのudpまでこの仕組みでOKなのかは確認していません。  

## 実装について  
１ソースファイルで実装するように心がけました。おおよそ３つのパートに分かれています。外部向け(export function/method)は、ソースコードの下のほうにしました。


## 参考  
golangの標準のnetのListenPacketの仕様 
[net - The Go Programming Language](https://golang.org/pkg/net/)  

```For UDP and IP networks, if the host in the address parameter is empty or a literal unspecified IP address, ListenPacket listens on all available IP addresses of the local system except multicast IP addresses.```
