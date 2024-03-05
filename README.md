# Starlit-Willow-Song

<h3>:sparkles: "柳星语" —— 交友、购物平台</h3>
<h4>

- 项目介绍
  </h4>

<i>
&emsp;&emsp;柳星语是一个交友、购物平台，集购物，发帖，对话于一体，<br>
&emsp;&emsp;采用了分布式架构和微服务设计理念，各个模块之间协同配合，业务上也做了不少的优化。<br>
&emsp;&emsp;技术栈： gin、gorm、grpc、mysql、redis、kafka<br>
&emsp;&emsp;功能：购物平台、论坛式交友平台
</i>

<br>

<h4>

- 项目亮点

</h4>
&emsp;1.&emsp;低耦合高内聚 + 异常自动重试<br>
&emsp;&emsp;<i>比如可以随时切换缓存配置，是使用了接口进行解耦</i><br>
&emsp;&emsp;<i>以及发现mysql或redis崩溃时，可以及时报警，并定时进行重连，无需重启服务</i><br><br>

&emsp;2.&emsp;采用三层限流设计<br>
&emsp;&emsp;<i>第一层是 使用nginx对 ip 进行限流访问</i><br>
&emsp;&emsp;<i>第二层是 网关服务对 用户id 进行限流访问</i><br>
&emsp;&emsp;<i>第三层是 微服务上 对每段时间的访问量做限制</i><br>

&emsp;3.&emsp;分布式系统和微服务框架<br>
&emsp;&emsp;<i>能够实现高可用性和可扩展性,以满足不断增长的需求<br>
&emsp;&emsp;以及更灵活地开发、部署和维护各个服务</i><br>

&emsp;4.&emsp;使用中间件 redis<br>
&emsp;&emsp;<i>作为缓存数据库，存储热点数据，加速数据访问速度，提高系统的性能</i><br>
&emsp;&emsp;<i>使用 redis做分布式锁，确保在多个实例之间的数据一致性和并发控制</i><br>
&emsp;&emsp;<i>使用 redis的原子操作和计数器功能，解决并发访问计数</i><br>

&emsp;5.&emsp;使用中间件 kafka<br>
&emsp;&emsp;<i>缓存请求并异步处理，平滑处理流量峰值，避免系统因高峰流量而崩溃</i><br>
&emsp;&emsp;<i>在不同的组件之间实现松耦合的通信，提高系统的可伸缩性和灵活性</i><br>

<h4>

- 性能测试

</h4>
&emsp;&emsp;<i>暂无数据</i>


