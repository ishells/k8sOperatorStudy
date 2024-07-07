/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (

	// 提供了TLS（Transport Layer Security）协议的支持，用于实现安全的网络通信，比如HTTPS。
	// 它可以帮助创建支持加密的网络连接，保护数据在传输过程中的安全。
	"crypto/tls"
	// 实现了命令行参数解析，允许程序从命令行输入中获取配置信息或开关控制程序行为。
	"flag"
	// 提供了与操作系统交互的功能，如文件操作、进程信号处理、环境变量访问等，是Go语言进行系统级编程的基础
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	// 初始化Kubernetes客户端认证插件。
	// 这些插件支持多种身份验证机制（如Azure、GCP、OIDC等），
	// 确保执行入口点和运行时能够利用它们进行Kubernetes集群的认证和授权。
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	// 提供了Kubernetes API对象的序列化和反序列化功能，以及类型转换和scheme注册机制，
	// 是处理Kubernetes资源对象的核心库。
	"k8s.io/apimachinery/pkg/runtime"
	// 包含了一些运行时实用函数，主要用于错误处理、日志记录和程序退出时的清理工作，
	// 帮助管理程序的运行时状态。
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	// 定义了Kubernetes API的各种Scheme，Scheme负责将API对象与其版本、组关联起来，
	// 是客户端与服务器间通信的基础。
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	// 用于简化编写Kubernetes控制器的流程。它提供了控制器生命周期管理、CRUD操作、事件处理、依赖注入等功能。
	ctrl "sigs.k8s.io/controller-runtime"
	// 提供健康检查功能，允许控制器暴露其健康状况给外部监控系统，是确保系统稳定性和可观察性的重要组件。
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	// 集成Zap日志库，为controller-runtime提供高性能、结构化的日志记录能力
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// 提供过滤器用于筛选metrics，有助于优化和控制导出到监控系统的指标数据。
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	// 用于启动一个HTTP服务器以供收集和导出metrics，支持Prometheus等监控系统采集。
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	// 提供webhook服务器实现，使得控制器可以作为webhook服务端，参与Kubernetes webhook机制，
	// 实现自定义的准入控制或其他扩展功能
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	batchv1 "tutorial.kubebuilder.io/project/api/v1"
	"tutorial.kubebuilder.io/project/internal/controller"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(batchv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	tlsOpts := []func(*tls.Config){}
	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	// 指定一组 controller 监听指定namespace的资源
	//var namespaces []string
	//defaultNamespaces := make(map[string]cache.Config)
	//
	//for _, ns := range namespaces {
	//	defaultNamespaces[ns] = cache.Config{}
	//}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// 指定一组 controller 监听指定namespace的资源
		//Cache: cache.Options{
		//	DefaultNamespaces: defaultNamespaces,
		//},

		// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
		// More info:
		// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/metrics/server
		// - https://book.kubebuilder.io/reference/metrics.html
		Metrics: metricsserver.Options{
			BindAddress:   metricsAddr,
			SecureServing: secureMetrics,
			// TODO(user): TLSOpts is used to allow configuring the TLS config used for the server. If certificates are
			// not provided, self-signed certificates will be generated by default. This option is not recommended for
			// production environments as self-signed certificates do not offer the same level of trust and security
			// as certificates issued by a trusted Certificate Authority (CA). The primary risk is potentially allowing
			// unauthorized access to sensitive metrics data. Consider replacing with CertDir, CertName, and KeyName
			// to provide certificates, ensuring the server communicates using trusted and secure certificates.
			TLSOpts: tlsOpts,
			// FilterProvider is used to protect the metrics endpoint with authn/authz.
			// These configurations ensure that only authorized users and service accounts
			// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
			// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/metrics/filters#WithAuthenticationAndAuthorization
			FilterProvider: filters.WithAuthenticationAndAuthorization,
		},
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "80807133.tutorial.kubebuilder.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controller.CronJobReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CronJob")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
