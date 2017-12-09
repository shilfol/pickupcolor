## 画像を読み込み特徴となる色を出力する
参考文献のgolang実装

## 使用方法
mainを走らせる
```
  git clone "this repository"
  go run imagefile      //引数なし 6色
  go run imagefile 7    //引数指定 7色

```

## 出力テスト
```
$ go run main.go yurufuwaplus.jpg 7
distance:  2028.382914985505
0 #7ba046
1 #d1e067
2 #319b7f
3 #e98c67
4 #a56f35
5 #eac110
6 #4cc4b2
$ ls
README.md  main.go  yurufuwaplus-pickupcolor7.png  yurufuwaplus.jpg 

```

#### 参考    
http://mrkn.hatenablog.com/entry/2015/07/12/000618
http://business.nikkeibp.co.jp/atclbdt/15/258678/071500002/?ST=print
