package registry

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/hashicorp/consul/api"
)

var (
	client    *api.Client
	serviceID string // 保存当前服务的ID，用于注销时使用
)

// getContainerIP 自动获取容器IP地址
func getContainerIP() string {
	// 方法1：从环境变量获取（优先级最高）
	if ip := os.Getenv("SERVICE_ADDR"); ip != "" {
		return ip
	}

	// 方法2：从配置文件获取（如果不为空）
	ctx := context.Background()
	if configIP := g.Cfg().MustGet(ctx, "consul.serviceAddr", "").String(); configIP != "" {
		return configIP
	}

	// 方法3：自动检测容器IP
	interfaces, err := net.Interfaces()
	if err != nil {
		g.Log().Errorf(ctx, "获取网络接口失败: %v", err)
		return "127.0.0.1"
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}

			// 优先返回172.19.0.x网段的IP（Docker网络）
			if ip[0] == 172 && ip[1] == 19 {
				g.Log().Infof(ctx, "自动检测到容器IP: %s", ip.String())
				return ip.String()
			}
		}
	}

	// 如果没有找到172.19.0.x网段的IP，返回第一个非回环IP
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip != nil {
				g.Log().Infof(ctx, "使用备选IP: %s", ip.String())
				return ip.String()
			}
		}
	}

	g.Log().Warningf(ctx, "无法获取容器IP，使用默认值: 127.0.0.1")
	return "127.0.0.1"
}

func InitConsul() *api.Client {
	ctx := context.Background()
	cfg := api.DefaultConfig()
	addr := g.Cfg().MustGet(ctx, "consul.addr").String()
	cfg.Address = addr

	var err error
	client, err = api.NewClient(cfg)
	if err != nil {
		g.Log().Fatalf(ctx, "init consul failed: %v", err)
	}

	return client
}

func RegisterService() error {
	ctx := context.Background()
	serviceName := g.Cfg().MustGet(ctx, "consul.serviceName").String()
	servicePort := g.Cfg().MustGet(ctx, "consul.servicePort").Int()

	// 自动获取服务地址
	serviceAddr := getContainerIP()

	// 生成固定的Service ID，基于服务名、IP和端口
	// 这样重启后会使用相同的ID，自动覆盖旧记录
	serviceID = fmt.Sprintf("%s-%s-%d", serviceName, serviceAddr, servicePort)

	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: serviceAddr,
		Port:    servicePort,
		Check: &api.AgentServiceCheck{
			TCP:                            fmt.Sprintf("%s:%d", serviceAddr, servicePort),
			Interval:                       "15s",
			Timeout:                        "10s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	if err := client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("register service failed: %v", err)
	}

	g.Log().Infof(ctx, "service registered: %s (ID: %s) at %s:%d", serviceName, serviceID, serviceAddr, servicePort)
	return nil
}

func DeregisterService() error {
	ctx := context.Background()

	if serviceID == "" {
		return fmt.Errorf("service ID is empty, service may not be registered")
	}

	if err := client.Agent().ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("deregister service failed: %v", err)
	}

	g.Log().Infof(ctx, "service deregistered: %s", serviceID)
	return nil
}
