# Pod高负载自动打印JAVA线程堆栈

## 背景

在实际的k8s维护场景中，我们时常会遇到 Pod CPU 突刺的现象。由于信息捕获依赖人工采集，有时甚至还未完成信息捕获，问题就已经恢复了，这让人非常头大。因此，我们需要一个可以自动捕获信息的工具。在GitHub上，我们发现了一个不错的项目k8s-java-debug-daemon，但由于年久失修，已无法满足当前环境的运行需求，所以我们决定对该项目进行二次开发。

## 优化内容
* 支持新版的Grafana
* 支持企业微信告警
* 支持k8s部署
* 支持configmap进行配置
* 修复n多已知问题

## 工作流程

与 Grafana 的告警联动，配合阿里的 arthas，来完成高CPU使用率线程的堆栈抓取。  
整体流程如下：

1. 为 Grafana 添加 webhook 类型的告警通知渠道，地址为该程序的 url(默认的hooks路径为 /hooks)。
2. 配置Grafana图表，并设置告警阈值
3. 当 webhook 触发时，程序会自动将 crawl.sh 脚本拷贝到对应 Pod 的容器中并执行。
4. 程序将 stdout 保存到本地文件，并推送到企业微信。

## 效果预览

* 企业微信告警展示

![](static/pic/1.png)

* “点击查看”跳转访问arthas捕获的线程堆栈

![](static/pic/2.png)

## 支持环境
Grafana v10.x (v9.x应该也支持，未测试)

## 配置说明
- config/config.yaml配置项

```yaml
server:
  port: 8099  # 服务监听端口
  maxNodeLockManager: 10 # 每node同时运行执行数为10
  domain: "http://127.0.0.1:8099" # 服务监听域名

wework:
  webhook: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxxxxxx" # 企业微信webhook地址
```
- 默认取最繁忙的前50个线程的堆栈信息 (可在 `crawl.sh` 中修改)
- 采集样本时间为2秒 (可在 `crawl.sh` 中修改)

## 如何使用

### Docker镜像

* 这是编译好的镜像，可以直接拉取使用

```text
docker pull registry.cn-hangzhou.aliyuncs.com/yilingyi/k8s-java-thread-dumper:2.0.1
```

* 自行构建

拉取源码
```text
git clone https://github.com/yilingyi/k8s-java-thread-dumper.git
```

构建镜像
```text
make docker IMAGE=yilingyi/k8s-java-thread-dumper:2.0.1
```

### Kubernetes部署

* 创建命名空间monitor

```text
kubectl create namespace monitor
```

* 将下面三个文件放在同一目录下，并使用`kubectl apply -f . -n monitor`进行k8s资源创建

Deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-java-thread-dumper
  labels:
    app: k8s-java-thread-dumper
spec:
  replicas: 1
  selector:
    matchLabels:
      app: k8s-java-thread-dumper
  template:
    metadata:
      labels:
        app: k8s-java-thread-dumper
    spec:
      containers:
      - name: k8s-java-thread-dumper
        image: registry.cn-hangzhou.aliyuncs.com/yilingyi/k8s-java-thread-dumper:2.0.1
        ports:
        - containerPort: 8099
        volumeMounts:
        - name: config-volume
          mountPath: /app/config/config.yaml
          subPath: config.yaml
      volumes:
      - name: config-volume
        configMap:
          name: k8s-java-thread-dumper-config
```

Service.yaml
```yaml
apiVersion: v1
kind: Service
metadata:
  name: k8s-java-thread-dumper-service
  labels:
    app: k8s-java-thread-dumper
spec:
  selector:
    app: k8s-java-thread-dumper
  ports:
  - protocol: TCP
    port: 8099
    targetPort: 8099
  type: NodePort
```

ConfigMap.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8s-java-thread-dumper-config
data:
  config.yaml: |
    server:
      port: 8099
      maxNodeLockManager: 10
      domain: "http://127.0.0.1:8099"
    wework:
      webhook: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxxxxxx"
```
* 资源授权

保存为rolebinding.yaml，并使用`kubectl apply -f rolebinding.yaml`进行创建，其中`<target-namespace>`改为目标命名空间

```text
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: <target-namespace>
  name: pod-exec-role
rules:
- apiGroups: [""]
  resources: ["pods/exec"]
  verbs: ["create"]

---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: pod-exec-role-binding
  namespace: <target-namespace>
subjects:
- kind: ServiceAccount
  name: default
  namespace: monitor
roleRef:
  kind: Role
  name: pod-exec-role
  apiGroup: rbac.authorization.k8s.io
```

#### Grafana告警规则

* Metrics Browser中填写

```text
sum(irate(container_cpu_usage_seconds_total{prometheus_name=~"gz",pod=~".*",container =~".*",container !="",container!="POD",node=~".*",namespace=~"(prod)"}[2m])) by (namespace, pod, node, container) / (sum(container_spec_cpu_quota{prometheus_name=~"gz",pod=~".*",container =~".*",container !="",container!="POD",node=~".*",namespace=~"(prod)"}/100000) by (namespace, pod, node, container)) * 100
```

Legend 中填写

```text
{{node}} - {{namespace}} - {{pod}} - {{container}}
```

配置完如下：
![](static/pic/3.png)

* 联络点配置

选择webhook,URL地址为http://xxxxx/hooks

配置完如下：
![](static/pic/4.png)

## **欢迎订阅我的公众号「SRE运维手记」**
![](static/pic/5.png)

参考链接
https://github.com/majian159/k8s-java-debug-daemon.git