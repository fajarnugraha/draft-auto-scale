# GCP Implementation of Auto-Scaling Requirements

This document describes how Google Cloud Platform (GCP) services, primarily the **Google Kubernetes Engine (GKE)**, can be used to meet the generic auto-scaling requirements.

GCP can fulfill all the specified requirements for auto-scaling containerized applications. GKE, a managed Kubernetes service, provides the core components for a robust and flexible auto-scaling strategy. For the most hands-off approach, **GKE Autopilot** mode manages the underlying nodes and cluster scaling automatically.

## 1. Application Scaling

### 1.1. Containerization

- **Container Support:** GKE is a conformant Kubernetes service, with native support for running industry-standard Docker containers.
- **Supported Runtimes:** GKE is language-agnostic. Any application that can be containerized, including **Go** and **JavaScript (Node.js)**, can be run and scaled on GKE.

### 1.2. Scaling Dimensions

GKE provides sophisticated mechanisms for horizontal scaling.

- **Horizontal Scaling:** This is achieved using the **Horizontal Pod Autoscaler (HPA)**. The HPA automatically adjusts the number of pod replicas in a Deployment or ReplicaSet based on observed metrics, directly fulfilling the requirement.

### 1.3. Scaling Triggers

GKE can trigger scaling events based on a wide variety of metrics.

- **Resource-Based Scaling:** The HPA natively supports scaling based on **CPU Utilization** and **Memory Utilization**. This is a standard and widely used configuration.

- **Request-Based Scaling:** The HPA can scale based on custom and external metrics.
    - **Request Count:** Metrics like requests-per-second (RPS) can be collected from the GKE Ingress controller and fed into the HPA via the Custom Metrics API, which is backed by **Google Cloud Monitoring**.
    - **Request Latency:** Similarly, latency metrics from the load balancer can be used as a custom metric to trigger scaling events.

- **Schedule-Based Scaling:** While Kubernetes does not have a native "Scheduled Pod Autoscaler," this is commonly implemented by using a Kubernetes **CronJob**. The CronJob runs at a specified schedule and executes a command (e.g., using `kubectl`) to patch the `minReplicas` and `maxReplicas` fields of the HPA object, effectively creating a scheduled scaling event.

- **Event-Driven Scaling:** This is achieved by integrating **KEDA (Kubernetes Event-driven Autoscaling)** with GKE. KEDA is an open-source component that can scale workloads based on metrics from various event sources, including **Google Cloud Pub/Sub**, and works by driving the HPA.

## 2. Scaling Performance

### 2.1. Time-to-Scale

The time-to-scale in GKE is a combination of Pod auto-scaling (HPA) and cluster-level scaling (Cluster Autoscaler).

- **Scale-Up Time:**
    1.  The HPA controller checks metrics every 15 seconds by default and reacts quickly to create new pods when a threshold is breached.
    2.  If there is not enough capacity on existing nodes, the **Cluster Autoscaler (CA)** is triggered to provision a new node.
    3.  The time to provision a new VM node, pull the container image, and start the container typically takes a few minutes. This process can be significantly accelerated by using **GKE Autopilot** or by configuring node pool settings and using smaller, optimized container images. The 5-minute requirement is generally achievable.

- **Scale-Down Time:** The HPA has a configurable stabilization window (defaulting to 5 minutes) to prevent rapid scale-down events. The Cluster Autoscaler will also safely drain and terminate underutilized nodes after a configurable period (typically 10 minutes). These settings can be tuned to meet specific cost and availability requirements.

### 2.2. Cooldown Periods

The HPA in GKE has built-in, configurable cooldown and stabilization periods. The `--horizontal-pod-autoscaler-downscale-stabilization` flag (configurable in the GKE control plane) allows for setting a specific cooldown period for scale-down events, preventing thrashing.

## 3. Configuration and Management

- **Declarative Configuration:** All GKE and Kubernetes resources, including Deployments and HPAs, are configured using declarative **YAML** files, fully supporting infrastructure-as-code (IaC) practices.
- **API and CLI Access:** GCP provides the **`gcloud` CLI** for managing GKE clusters, and Kubernetes provides the **`kubectl` CLI** for managing workloads within the cluster. Both are backed by comprehensive REST APIs.
- **Monitoring and Logging:** **Google Cloud's operations suite (formerly Stackdriver)** provides deep, out-of-the-box integration with GKE. It automatically collects logs, events, and metrics, including detailed information on all auto-scaling activities, for comprehensive monitoring and auditing.

## 4. Case Studies

### 4.1. Online Shop Flash Sale
The implementation for the flash sale case study on GKE is identical to the generic Kubernetes implementation. It uses a combination of `CronJob` resources for proactive scheduled scaling and the standard `HorizontalPodAutoscaler` for reactive scaling during the event.

Refer to the [generic Kubernetes implementation](./K8S.md#4.1-online-shop-flash-sale) for the detailed steps.

### 4.2. Online Test Platform

This use case requires a more sophisticated setup than the flash sale due to its stateful nature and the need to handle unpredictable, tenant-driven events.

### Implementation Steps

1.  **Session Persistence (Sticky Sessions):**
    -   This is a critical prerequisite. On GKE, this is configured on the **BackendConfig** resource associated with the service.
    -   You would set the `affinityType` to `CLIENT_IP` for simple affinity or `GENERATED_COOKIE` for a more robust cookie-based approach. This ensures the Google Cloud Load Balancer always directs a specific user to the same pod.

2.  **Scenario 1: Predictable, Coordinated Event:**
    -   The implementation is identical to the generic Kubernetes use case. A **CronJob** is used to proactively patch the HPA resource with a higher `minReplicas` count before the event begins, and a second CronJob scales it back down afterward.

3.  **Scenario 2: Unpredictable, Tenant-Driven Event:**
    -   This requires an event-driven approach using **KEDA** combined with a GCP service like **Cloud Pub/Sub**.
    -   **Application Change:** The application must be modified to publish an event to a Cloud Pub/Sub topic when a teacher schedules an exam. The event payload should contain the number of students and the start time.
    -   **KEDA Scaler:** A KEDA `ScaledObject` is created and configured with the `gcp-pubsub` scaler, pointing to the appropriate subscription. KEDA will monitor the number of messages in the subscription and scale the pods accordingly. For more advanced logic (e.g., scaling based on a value in the message payload), a custom metrics adapter service would be used, as described in the generic Kubernetes implementation.

4.  **Graceful, Session-Aware Scale-Down:**
    -   The implementation is identical to the generic Kubernetes approach.
    -   The application pods must be enhanced with more intelligent readiness probes and a `preStop` lifecycle hook.
    -   The readiness probe should fail if the pod has active sessions, preventing the GKE Ingress from sending it new traffic.
    -   The `preStop` hook gives the pod time to wait for existing sessions to complete before Kubernetes terminates it. This ensures no student is cut off mid-exam during a scale-down event.

## 5. Testing and Monitoring

### 5.1. Load Generation
To validate the HPA and Cluster Autoscaler configurations on GKE, a load generator is essential. A modern tool like **k6** is recommended.

Refer to the guide on [Load Testing with k6](../load-generator/k6.md) for a detailed example of how to create a test script and generate traffic against your GKE service.

### 5.2. Monitoring Dashboard
GKE is deeply integrated with **Google Cloud's operations suite** (formerly Stackdriver), which provides powerful monitoring and dashboarding capabilities out of the box.

A recommended dashboard in **Cloud Monitoring** for observing GKE auto-scaling would be configured with these widgets:
-   **HPA Target vs. Current Metric:** A chart showing the current average CPU utilization from the GKE container metrics versus the HPA's target utilization.
-   **Pod Count (Replicas):** A chart displaying the number of available pods in the deployment, which directly visualizes the result of HPA's scaling decisions.
-   **Requests Per Second (RPS):** If scaling on traffic, a chart showing the RPS metric from the Cloud Load Balancer.
-   **Node Count:** A chart showing the number of nodes in the GKE cluster's node pool, which visualizes the GKE Cluster Autoscaler's activity.

These metrics are available by default when GKE monitoring is enabled, making it straightforward to build a comprehensive dashboard.
