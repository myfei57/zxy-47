# 基于 Go 实现的 GreenHouse 项目，一款后端服务，完成温室大棚的遮阳/侧窗/卷膜/风机联动、墒情灌溉与水肥配比、生育期 EC 目标、水肥配额与补光分区的环境群控。

GreenHouse 面向设施农业温室集群，按棚分区管理环境：光照与墒情探头持续采样落盘，
气候模式守卫按风向与阈值联动遮阳帘、上下风侧窗、卷膜机与环流风机；水肥机按生育期
目标先兑水再加母液并稳定 pH 后校正 EC；灌溉计划按实时墒情执行并受每日配额约束；
补光区段随分区调整刷新传感器归属；全部执行与告警进入审计。

## 构建与运行

```bash
go build -mod=vendor -o greenhouse.exe ./cmd/greenhouse
./greenhouse.exe -data data -listen :8080
```

访问 http://127.0.0.1:8080/sheds 、/climate 、/irrig 、/alarms 查看四个控制台页面，
健康检查位于 /healthz，控制与查询接口位于 /api。

## Docker

```bash
bash build_benzhi_docker.sh greenhouse linux/amd64
docker run --rm -p 8080:8080 greenhouse -listen :8080
```
