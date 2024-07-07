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

package v1

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CronJobSpec defines the desired state of CronJob
type CronJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CronJob. Edit cronjob_types.go to remove/update
	// Foo string `json:"foo,omitempty"`

	// +kubebuilder:validation:MinLength=0
	// The schedule in Cron format
	Schedule string `json:"schedule"`

	// +kubebuilder:validation:Minimum=0
	// 如果错过计划，开始作业的可选截止日期，以秒为单位时间
	// 无论出于什么原因，错过的作业执行将被视为失败。
	// +optional
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty"`

	// 指定如何处理作业的并发执行。
	// 有效值是：
	// -“Allow”（默认）：允许CronJobs同时运行；
	// -“Forbid”：禁止并发运行，如果上一个运行尚未完成，则跳过下一次运行；
	// -“Replace”：取消当前正在运行的作业，并将其替换为新作业
	// +optional
	ConcurrencyPolicy ConcurrencyPolicy `json:"concurrencyPolicy,omitempty"`

	// 此标志告诉控制器暂停后续执行，
	// 已经开始的执行 不能使用该字段。默认值为false。
	// +optional
	Suspend *bool `json:"suspend,omitempty"`

	// 指定执行CronJob时将创建的作业。
	JobTemplate batchv1.JobTemplateSpec `json:"jobTemplate"`

	// +kubebuilder:validation:Minimum=0
	// 指定要保留的执行成功的Job的数量。
	// 这是一个区分显式零和未指定的指针
	// +optional
	SuccessfulJobHistoryLimit *int32 `json:"successfulJobHistoryLimit,omitempty"`

	// +kubebuilder:validation:Minimum=0
	// 指定要保留的执行失败的Job的数量。
	// +optional
	FailedJobHistoryLimit *int32 `json:"failedJobHistoryLimit,omitempty"`
}

// ConcurrencyPolicy 定义一个自定义类型来保存我们的并发策略。
// 它实际上只是一个字符串，但该类型提供了额外的文档，
// 并允许我们在类型而不是字段上附加验证，使验证更容易重用。
// ConcurrencyPolicy描述了如何处理作业。
// 只能指定以下并发策略之一。
// 如果未指定以下策略，则默认策略是AllowConcurrent
// +kubebuilder:validation:Enum=Allow;Forbid;Replace
type ConcurrencyPolicy string

const (
	// AllowConcurrent allows CronJobs to run concurrently.
	AllowConcurrent ConcurrencyPolicy = "Allow"

	// ForbidConcurrent forbids concurrent runs, skipping next run if previous
	// hasn't finished yet.
	ForbidConcurrent ConcurrencyPolicy = "Forbid"

	// ReplaceConcurrent cancels currently running job and replaces it with a new one.
	ReplaceConcurrent ConcurrencyPolicy = "Replace"
)

// CronJobStatus defines the observed state of CronJob
type CronJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// 指向当前运行作业的指针列表。
	// +optional
	Active []corev1.ObjectReference `json:"active,omitempty"`

	// +optional
	// job上次被成功调度的信息
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CronJob is the Schema for the cronjobs API
type CronJob struct {
	// Root Object Definitions
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronJobSpec   `json:"spec,omitempty"`
	Status CronJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CronJobList contains a list of CronJob
type CronJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CronJob{}, &CronJobList{})
}

/*
在这份Go代码中，以 // + 开头的注释具有特殊意义，它们是Kubernetes API Machinery的标记，
主要用于自动生成CRD（CustomResourceDefinition）配置和其他元数据。
这些标记帮助开发者通过简单的注释来指定API对象的行为、验证规则等，而无需手动编写大量 YAML 或 JSON 配置。

下面是一些例子及其含义：
Validation Markers:
// +kubebuilder:validation:MinLength=0: 表示该字段的最小长度为0，用于字符串类型的字段验证。
// +kubebuilder:validation:Minimum=0: 表示该字段的最小值为0，用于数值类型的字段验证。
// +kubebuilder:validation:Enum=Allow;Forbid;Replace: 限制字段的取值只能是给定枚举中的一个，这里用于定义ConcurrencyPolicy的合法取值。

Optional Field Marker:
// +optional: 标记该字段为可选，意味着在序列化时如果字段值为零值（如nil、空字符串等），则不会出现在生成的API对象中。

Root Type Marker:
// +kubebuilder:object:root=true: 标记该结构体为API资源的根类型，通常用于CRD的定义。

Subresource Marker:
// +kubebuilder:subresource:status: 表示该资源有一个子资源是status，允许单独更新资源的状态部分。

Enum Type Marker:
虽然没有直接的例子，但在自定义类型如ConcurrencyPolicy的上下文中，这些注释虽然不以// +开头，但它们定义了类型并间接通过其他地方的验证注解关联，确保字段值的合法性。
这些标记极大简化了Kubernetes自定义资源开发过程，通过代码注释即可实现复杂的API元数据定义和验证逻辑。
*/

/*

Kubernetes API Machinery是Kubernetes架构中的一个核心组件，它提供了一套框架用于构建高效、可靠且可扩展的API服务器。
API Machinery负责处理API请求的接收、解析、验证、处理以及响应，是Kubernetes控制面与其他组件（如kubectl、kubelet、operators等）以及外部系统之间通信的中枢。

API Machinery的关键组成部分包括：
① 序列化/反序列化: 提供了将HTTP请求和响应转换为Kubernetes对象（如Pods、Services）的能力，以及反之亦然。这主要通过Protobuf或JSON格式完成。
② 存储层: 抽象了后端数据存储，使得API服务器可以使用多种数据存储解决方案（如etcd）而无需修改上层逻辑。
③ 插件机制: 允许插入各种插件，如自定义的认证、授权、准入控制插件，以及各种其他的扩展点，增强了API服务器的功能性和安全性。
④ 工作队列和控制器: 支持异步处理和长期运行的操作，如通过Operator模式管理自定义资源（CRDs）。
⑤ API组、版本和资源: 提供了灵活的API分组、版本管理和资源注册机制，使得Kubernetes可以轻松扩展新的APIs，同时保持向后兼容性。
⑥ OpenAPI和Swagger: 自动生成API文档，便于开发者和自动化工具理解可用的API接口。
⑦ 缓存和性能优化: 包括索引、观看机制和高效的数据分发，以支持大量客户端的实时更新需求。

简而言之，API Machinery是构建和扩展Kubernetes API的基础，它不仅支撑着Kubernetes内建资源的操作，
也是自定义资源（CRDs）、聚合API服务器等高级特性的技术基石。开发者可以通过理解和利用API Machinery，来定制和扩展Kubernetes以满足特定应用场景的需求。
*/
