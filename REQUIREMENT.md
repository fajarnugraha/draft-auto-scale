# Generic Requirements for Cloud Provider Auto Scaling

This document outlines the generic requirements for a cloud provider to support auto-scaling features for containerized applications. The requirements are based on existing features in major cloud providers and are intended to be cloud-agnostic.

## 1. Application Scaling

The primary focus of these requirements is on application-level scaling using containers.

### 1.1. Containerization

- **Container Support:** The platform must support running applications in industry-standard containers (e.g., Docker).
- **Supported Runtimes:** The platform should have first-class support for popular programming language runtimes, including **Go** and **JavaScript (Node.js)**.

### 1.2. Scaling Dimensions

The auto-scaling mechanism should support horizontal scaling.

- **Horizontal Scaling:**
    - The platform must be able to automatically increase or decrease the number of container instances (replicas) based on demand.
    - This is the primary mechanism for scaling stateless applications.

### 1.3. Scaling Triggers

The auto-scaling system should be able to trigger scaling events based on a variety of metrics.

- **Resource-Based Scaling:**
    - **CPU Utilization:** Scale based on the average CPU utilization of all running container instances.
    - **Memory Utilization:** Scale based on the average memory utilization of all running container instances.

- **Request-Based Scaling:**
    - **Request Count:** Scale based on the number of incoming requests per second (RPS) to a load balancer or ingress controller.
    - **Request Latency:** Scale based on the average response time of the application.

- **Schedule-Based Scaling:**
    - The platform must allow for scheduled scaling events to handle predictable traffic patterns (e.g., scaling up during business hours and down at night).

- **Event-Driven Scaling:**
    - The platform should support scaling based on events from other services, such as the number of messages in a queue or tasks in a stream.

## 2. Scaling Performance

The auto-scaling system must be able to react quickly to changes in load to ensure application availability and performance.

### 2.1. Time-to-Scale

- **Scale-Up Time:** When the average load exceeds a predefined threshold (e.g., 70% CPU utilization) for a configurable period (e.g., 60 seconds), the auto-scaling system must complete the scale-up operation (i.e., have new container instances ready to serve traffic) within a specified timeframe (e.g., 5 minutes). The goal is to bring the load back below a target threshold (e.g., 50% CPU utilization).

- **Scale-Down Time:** When the average load falls below a predefined threshold (e.g., 30% CPU utilization) for a configurable period (e.g., 300 seconds), the auto-scaling system should gracefully terminate excess container instances. This process should be completed within a specified timeframe (e.g., 5 minutes) to reduce costs, without impacting application availability.

### 2.2. Cooldown Periods

- To prevent rapid fluctuations in the number of instances, the auto-scaling system must support configurable cooldown periods between scaling activities. This prevents the system from launching or terminating instances before the previous scaling activity has had a chance to take effect.

## 3. Configuration and Management

- **Declarative Configuration:** The auto-scaling policies should be definable in a declarative format (e.g., YAML, JSON) to enable infrastructure-as-code practices.
- **API and CLI Access:** The platform must provide a comprehensive API and command-line interface (CLI) for managing auto-scaling configurations.
- **Monitoring and Logging:** The platform must provide detailed monitoring and logging of auto-scaling events to allow for auditing and troubleshooting.

## 4. Case Examples

### 4.1. Online Shop Flash Sale

This use case describes an online retail store preparing for a flash sale event.

- **Scenario:** A flash sale is scheduled to start at a specific time (e.g., Friday at 8:00 PM) and last for one hour. The expected traffic is 100 times the normal load.
- **Requirement:** The system must handle the massive, sudden surge in traffic without impacting user experience. Customers should not experience slow loading times or be unable to access the site.

- **Implementation Requirements:**
    - **Proactive Scheduled Scaling:** The platform must support a scheduled scaling mechanism to pre-warm the environment. The number of container instances should be significantly increased *before* the sale begins to handle the anticipated load.
    - **Reactive Scaling during the Event:** During the sale, the system must still be able to react to unexpected traffic spikes.
        - **Threshold:** A reactive scale-up should be triggered if the average CPU utilization exceeds a sensible threshold (e.g., 75%).
        - **Time-to-React:** The time from when the threshold is breached to when new container instances are ready to serve traffic must be very short (e.g., under 2 minutes) to prevent service degradation.
        - **Scheduled Scale-Down:** To optimize costs, the system must automatically scale down to normal levels after the event is over.

- **Special Considerations:**
    - **Threshold Rationale (75%):** This threshold provides a good balance. It indicates that the system is under significant load but leaves a 25% buffer to absorb momentary traffic spikes without immediately dropping requests, giving the autoscaler time to react.
    - **Time-to-React Rationale (2 minutes):** For a high-stakes, short-lived event like a flash sale, a rapid response is critical. A 2-minute window from detection to readiness is an aggressive but necessary goal to prevent user abandonment. It forces the use of optimized container images and a cluster configured for fast node provisioning.
    - **Scale-Down Rationale (Scheduled):** A reactive scale-down is not used because the event has a predictable end time. A scheduled scale-down is more efficient, as it can terminate the expensive, high-capacity resources immediately after the sale concludes, rather than waiting for traffic to naturally decrease and a cooldown period to pass.

## 5. Testing and Monitoring

To ensure the auto-scaling system is reliable and performs as expected, the platform must provide robust testing and monitoring capabilities.

### 5.1. Load Generation

- **Requirement:** A load generation tool must be used to simulate various traffic patterns (e.g., sudden spikes, gradual increases) to validate that the auto-scaling triggers and policies work correctly.
- **Implementation:** The implementation guides should refer to a specific, modern load generation tool. See `implementation/load-generator/` for examples.

### 5.2. Monitoring Dashboard

- **Requirement:** A real-time monitoring dashboard is required to visualize the state of the auto-scaling system.
- **Key Metrics:** The dashboard must display, at a minimum:
    - The number of active container instances (pods/replicas).
    - The current value of the metric driving the scaling decision (e.g., average CPU utilization, requests per second).
    - The scaling metric's target threshold.
    - Historical data for these metrics to analyze scaling behavior over time.

### 4.2. Online Test Platform

This use case describes a multi-tenant online testing platform where individual tenants (e.g., schools, universities) can schedule their own exams. This presents two distinct scaling challenges.

-   **Scenario 1 (Predictable):** A large-scale, coordinated event, like a national standardized test. The platform provider is aware of the event schedule in advance.
-   **Scenario 2 (Unpredictable):** A single tenant, like a university professor, schedules an exam for their class of 500 students to begin in one hour, without notifying the platform provider.

-   **Requirement:** The system must provide a seamless and fair experience for all students, regardless of whether the event was planned by the provider or a tenant. Student work must not be lost, and session integrity must be maintained.

-   **Implementation Requirements:**
    -   **Scenario 1 (Predictable):** Must support **proactive, scheduled scaling** to have capacity ready before the known event begins. This is identical to the flash sale use case.
    -   **Scenario 2 (Unpredictable):** Must support **event-driven scaling**. The system cannot rely on lagging indicators like CPU load. Instead, it must scale based on leading indicators from the application itself. For example, when an exam is created for a large number of students, the application should publish an `exam_scheduled` event, which an event-driven autoscaler can use to proactively scale up the necessary resources just in time.

-   **Special Considerations:**
    -   **Session Persistence (Sticky Sessions):** This is the most critical difference from a stateless e-commerce site. A student's session is stateful. The platform's load balancer must be configured to ensure that a student is always routed to the same container instance for the entire duration of their test to avoid losing progress.
    -   **Graceful, Session-Aware Scale-Down:** The system cannot simply terminate instances when a scheduled time ends. The scale-down process must be intelligent. It must not terminate an instance until it has verified that all active test sessions on that instance have been completed and submitted successfully. This may require a significant delay between the test's end time and the actual scaling down of resources.
