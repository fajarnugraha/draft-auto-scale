# Generic Kubernetes Implementation of Auto-Scaling Requirements

This document outlines how a generic, cloud-agnostic Kubernetes cluster can implement the auto-scaling requirements.

A standard Kubernetes installation provides the foundational components needed to create a powerful auto-scaling system. The implementation relies on core Kubernetes objects and some widely adopted open-source components.

## 1. Application Scaling

### 1.1. Containerization

- **Container Support:** Kubernetes is the de-facto standard for container orchestration and uses the Container Runtime Interface (CRI) to work with various container runtimes, including Docker.
- **Supported Runtimes:** As a container orchestration platform, Kubernetes is language-agnostic. It can run any containerized application, including those built with **Go** and **JavaScript (Node.js)**.

### 1.2. Scaling Dimensions

- **Horizontal Scaling:** This is a core feature of Kubernetes, implemented via the **Horizontal Pod Autoscaler (HPA)**. The HPA automatically scales the number of pods in a controller (like a Deployment or StatefulSet) based on observed metrics.

### 1.3. Scaling Triggers

Kubernetes can scale workloads based on a variety of metrics, though some require additional components.

- **Resource-Based Scaling:**
    - To enable scaling on **CPU Utilization** and **Memory Utilization**, the **Kubernetes Metrics Server** must be installed in the cluster. This lightweight component collects resource metrics from each node's Kubelet and exposes them via the Metrics API, which the HPA uses to make scaling decisions.

- **Request-Based Scaling:**
    - **Request Count & Latency:** To scale on metrics like RPS or latency, a more advanced monitoring solution is required. A common approach is to use **Prometheus** to scrape metrics from an Ingress Controller (like NGINX or Traefik) and the **Prometheus Adapter for Kubernetes Custom Metrics**. This adapter exposes the scraped metrics to the HPA via the Custom Metrics API, allowing scaling decisions based on request-level data.

- **Schedule-Based Scaling:**
    - This can be achieved using a standard Kubernetes **CronJob**. The CronJob can be scheduled to run at any desired time and can execute a `kubectl` command to patch the `minReplicas` and `maxReplicas` values of an HPA object, effectively creating a scheduled scaling policy.

- **Event-Driven Scaling:**
    - The most common and robust solution for event-driven scaling in Kubernetes is **KEDA (Kubernetes Event-driven Autoscaling)**. KEDA is an open-source component that integrates with the HPA to provide scaling based on events from dozens of sources, such as message queues (RabbitMQ, Kafka), databases, and cloud services.

## 2. Scaling Performance

### 2.1. Time-to-Scale

The time it takes to scale in Kubernetes depends on both pod scaling (HPA) and the underlying node scaling.

- **Scale-Up Time:**
    1.  The HPA controller checks metrics every 15 seconds by default.
    2.  If new pods are created and there isn't enough capacity on the existing nodes, the pods will remain in a `Pending` state.
    3.  To automatically add new nodes to the cluster, a **Cluster Autoscaler** component must be installed. This component is specific to the underlying infrastructure (e.g., the AWS Cluster Autoscaler for a cluster on AWS).
    4.  The time to provision a new node and for it to join the cluster is cloud-provider dependent but typically takes a few minutes. The 5-minute requirement is generally achievable with a properly configured Cluster Autoscaler.

- **Scale-Down Time:** The HPA has a configurable stabilization window (defaulting to 5 minutes for scale-downs) to prevent thrashing. The Cluster Autoscaler will also safely drain and terminate underutilized nodes after a configurable period (typically 10 minutes).

### 2.2. Cooldown Periods

The Kubernetes HPA has built-in cooldown/stabilization logic. The `--horizontal-pod-autoscaler-downscale-stabilization` and `--horizontal-pod-autoscaler-upscale-stabilization` flags on the controller-manager allow for fine-tuning these periods to prevent rapid, repeated scaling events.

## 3. Configuration and Management

- **Declarative Configuration:** All Kubernetes objects, including Deployments, HPAs, and CronJobs, are defined using declarative **YAML** manifests. This is a core principle of Kubernetes and fully enables GitOps and Infrastructure-as-Code (IaC) workflows.
- **API and CLI Access:** The primary way to interact with a Kubernetes cluster is via the **`kubectl` CLI** and the **Kubernetes API**, which provide comprehensive control over all cluster resources.
- **Monitoring and Logging:** While Kubernetes itself provides basic logging via `kubectl logs`, a complete monitoring and logging solution typically involves open-source tools like **Prometheus** for metrics, **Grafana** for visualization, and a logging stack like **Fluentd**, **Elasticsearch**, and **Kibana (EFK)**. These tools provide the necessary visibility into auto-scaling events and application behavior.

## 4. Case Study: Online Shop Flash Sale

This section details how to implement the flash sale use case using Kubernetes.

- **Normal Operation:**
    - An HPA is configured for the web application's Deployment.
    - `minReplicas: 5`
    - `maxReplicas: 20`
    - `targetCPUUtilizationPercentage: 60`

### Implementation Steps

1.  **Proactive Scheduled Scaling (Scale-Up):**
    - A **CronJob** is created to pre-warm the environment. It is scheduled to run 30 minutes before the sale starts (e.g., Friday at 7:30 PM).
    - The CronJob executes a `kubectl patch` command to update the HPA resource.
    - It modifies the HPA to handle the expected 100x load, setting a much higher baseline for the number of pods.
        - `minReplicas: 400`  *(This ensures a large number of pods are ready immediately)*
        - `maxReplicas: 800`  *(This provides headroom for unexpected surges beyond the estimate)*

2.  **Reactive Scaling During the Event:**
    - The existing HPA continues to function during the sale, but with the new replica counts.
    - If the CPU load across the 400+ pods still exceeds the `targetCPUUtilizationPercentage` (e.g., 60%), the HPA will automatically scale out further, up to the new maximum of 800 pods.
    - **Cluster Scaling:** A properly configured **Cluster Autoscaler** is crucial. It will see the large number of `Pending` pods created by the CronJob and HPA and will start provisioning new nodes in the cluster to accommodate them. This node scaling must complete before the sale begins.

3.  **Scheduled Scale-Down:**
    - A second **CronJob** is created to run after the sale is over (e.g., Friday at 9:05 PM).
    - This job patches the HPA again, returning its values to the normal, non-sale configuration.
        - `minReplicas: 5`
        - `maxReplicas: 20`
    - The HPA will then begin to scale the application down to the normal operational level. The Cluster Autoscaler will subsequently see the underutilized nodes and terminate them to reduce costs.

## 5. Testing and Monitoring

### 5.1. Load Generation

To validate the HPA and Cluster Autoscaler configurations, a load generator is essential. A modern tool like **k6** is recommended.

Refer to the guide on [Load Testing with k6](../load-generator/k6.md) for a detailed example of how to create a test script and generate traffic.

### 5.2. Monitoring Dashboard

For a generic Kubernetes cluster, the standard for monitoring is the combination of **Prometheus** and **Grafana**.

-   **Prometheus:** An open-source monitoring system that scrapes metrics from configured endpoints. It should be configured to scrape metrics from:
    -   The **Kubernetes Metrics Server** (for CPU/Memory).
    -   An **Ingress Controller** (for RPS/latency).
    -   The applications themselves (if they expose custom metrics).
-   **Grafana:** An open-source visualization tool used to create dashboards from data sources like Prometheus.

A recommended Grafana dashboard for monitoring auto-scaling would include these panels:
-   **HPA Target Metric vs. Current Metric:** A graph showing the current average CPU utilization across all pods versus the target utilization (e.g., 60%). This is the most critical view to see *why* scaling is happening.
-   **Pod Count (Replicas):** A graph showing the desired number of replicas set by the HPA versus the actual number of running replicas.
-   **Requests Per Second (RPS):** If scaling on a request-based metric, a graph showing the RPS from the Ingress controller.
-   **Cluster Node Count:** A graph showing the number of nodes in the cluster, which helps visualize the Cluster Autoscaler's activity.

## 6. Case Study: Online Test Platform

This use case requires a more sophisticated setup than the flash sale due to its stateful nature and the need to handle unpredictable, tenant-driven events.

### Implementation Steps

1.  **Session Persistence (Sticky Sessions):**
    -   This is a critical prerequisite. The **Ingress Controller** (e.g., NGINX, Traefik, HAProxy) must be configured to provide session affinity. This is often done using a cookie-based approach. For example, the Ingress object would be annotated to instruct the controller to set a cookie on the user's first visit and then use that cookie to route all subsequent requests to the same pod.

2.  **Scenario 1: Predictable, Coordinated Event:**
    -   The implementation is identical to the flash sale use case. A **CronJob** is used to proactively patch the HPA resource with a higher `minReplicas` count before the event begins, and a second CronJob scales it back down afterward.

3.  **Scenario 2: Unpredictable, Tenant-Driven Event:**
    -   This requires an event-driven approach using **KEDA**.
    -   **Application Change:** The application must be modified to publish an event to a message broker (e.g., RabbitMQ, Kafka) when a teacher schedules an exam. The event payload should contain the number of students and the start time.
    -   **KEDA Scaler:** A KEDA `ScaledObject` is created. Instead of monitoring CPU, it is configured with a scaler for the message broker (e.g., a `rabbitmq` scaler).
    -   **Custom Logic:** A separate small service (sometimes called a "metrics adapter" or "scaler service") is needed. This service consumes the `exam_scheduled` events and exposes a simple metric that KEDA can poll (e.g., `active_exams` or `pending_student_count`).
    -   **Scaling Action:** KEDA polls this service. When it sees that a large exam is scheduled to start soon, it drives the HPA to scale up the number of pods proactively, just in time for the event.

4.  **Graceful, Session-Aware Scale-Down:**
    -   This is the most complex part. The default HPA behavior (scaling down after a 5-minute stabilization window) is not sufficient as it could terminate a pod with an active student session.
    -   **Application Health Checks:** The application pods must be enhanced with more intelligent health checks. The pod should report itself as "unhealthy" or "unready" if it has active test sessions, even if the test's official end time has passed.
    -   **Pre-Stop Hook:** A `preStop` lifecycle hook should be configured for the container. This hook would trigger a script that waits for all active sessions to complete before allowing the pod to be terminated. This prevents Kubernetes from forcefully killing a pod while a student is still submitting their exam.
    -   By combining these techniques, the HPA can safely scale down the application, as pods will only be terminated after they have confirmed that all user sessions within them are complete.
