# · 难点日记

### · &emsp;此处记录我遇到的痛点和难点，其实也是吐糟一下自己还得练

<h3>&emsp;· 1. 使用空间索引时遇到的问题</h3>
<h3><i>&emsp;&emsp;· (1.1.1). 问题描述</i></h3>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;我自己创建了一个point类，但是使用gorm封装好的Create方法来创建记录时</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;他总是会自动的为其加上单引号，什么意思呢？举个例子...</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;比如我要添加的是 POINT(0,0)，正常写的话应该是 insert into table_name values (...,
POINT(0,0), ...);</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;但是gorm的Create方法却总是写成 insert into table_name values (..., 'POINT(0,0)', ...)
，总是多了一对单引号</h4>

<h3><i>&emsp;&emsp;· (1.1.2). 解决方案</i></h3>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;所以我后来就只好把关于POINT的方法都写成的 sql语句...</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;我不知道是gorm的问题，还是我没有完全熟练的使用gorm ...</h4>

<h3><i>&emsp;&emsp;· (1.2.1). 问题描述</i></h3>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;既然创建都无法创建了，那我想要读取POINT类型的字段就更不可能了 ...</h4>
<h4>
&emsp;&emsp;&emsp;&emsp;&emsp;因为我在go中找不到一个可以接收POINT类型的类型，即使我自己写一个Point结构体，也无济于事...</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;结构体代码:</h4>

```go
        // Point 定义地理位置的结构体
        type Point struct {
            X float64 //	经度
            Y float64 //	纬度
        }
```

<h3><i>&emsp;&emsp;· (1.2.2). 解决方案 (尚未彻底解决...)</i></h3>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;我也尝试过很多办法，看过很多帖子，比如说什么使用string类型或者[]byte类型来接收，</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;又或者是，重写结构体的Scan等方法，但是都没什么效果</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;但最后我发现，我的业务好像不用查询POINT，因为可以在Mysql中计算好后，再将对应的记录导出</h4>


<h3>&emsp;· 2. 在解散群聊遇到的问题</h3>
<h3><i>&emsp;&emsp;· (2.1). 问题描述</i></h3>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;我在用户信息表中，保存了一个用户所在群的一个字符串列表，例如: "1,2,3",表示该用户加入了id为1,2,3的群聊</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;而当群主退出群聊时，这个群应该被解散，我应该把这个群的记录删除</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;但是我还应该在用户的所在群列表中删除这个群id</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;正常的话是，是先查出一个用户的所在群列表，然后进行更改后再提交回 mysql，</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;但是一个群的上限我设置为500人，如果只是单个修改，这对性能而言，明显有着巨大的问题...</h4>

<h3><i>&emsp;&emsp;· (2.2). 解决方案</i></h3>
<h3>&emsp;&emsp;&emsp;&emsp;&emsp;· (2.2.1).</h3>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;通过添加一个status的字段标记群的状态，如果某个群被解散了，就将其标记为deleted</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;然后通过定时任务，在每天凌晨时间段(因为人少，数据库的压力小) 来进行记录的彻底删除</h4>

<h3>&emsp;&emsp;&emsp;&emsp;&emsp;· (2.2.2).</h3>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;我想到了mysql内置的字符串操作，使用 Replace方法来在mysql中直接修改，减少服务器和mysql之间的操作</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;但是原本的1，2，3这样的列表很难用Replace操作，因为没法准确的定位一个id</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;所以我只能将字符串列表的格式改为: ,1,2,3,</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;这样我就可以通过检验 ",*,"的方式锁定我需要删除的id了 ...</h4>
<h4>&emsp;&emsp;&emsp;&emsp;&emsp;可悲之处是，我要修改我以前写好的一些接口，所以下次我应该会直接使用后一种方式了...</h4>