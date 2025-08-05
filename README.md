# Auto-Scaling Requirements and Implementation Patterns

## 1. Overview

This repository contains a collection of documents defining a generic set of requirements for a cloud-native auto-scaling solution. It also provides practical implementation guides and case studies for various platforms.

The goal is to establish a clear, well-documented baseline for designing and implementing robust auto-scaling systems.

## 2. Repository Structure

The repository is organized as follows:

-   `REQUIREMENT.md`: The core document outlining the generic requirements for any auto-scaling system.
-   `/implementation`: Contains practical guides for implementing the requirements on different platforms.
    -   `/cloud-provider`: Platform-specific guides (Kubernetes, GCP, etc.).
    -   `/load-generator`: Guides for load testing tools.
-   `/summary`: Contains session summaries for project context and resumption.
-   `/prompts`: Contains a history of prompts used to generate the content in this repository.

## 3. Key Documents

-   **[Generic Requirements](./REQUIREMENT.md):** Start here to understand the core requirements for auto-scaling.
-   **Implementation Guides:**
    -   **[Kubernetes](./implementation/cloud-provider/K8S.md):** A guide for a generic Kubernetes cluster using open-source tools.
    -   **[Google Cloud (GKE)](./implementation/cloud-provider/GCP.md):** A guide for implementing the requirements on Google Kubernetes Engine.
-   **Tooling:**
    -   **[Load Testing with k6](./implementation/load-generator/k6.md):** A guide to using the k6 tool for load testing the auto-scaling setup.

## 4. How to Use

These documents are intended to be a living reference. They can be used as a starting point for designing new auto-scaling systems or as a benchmark for evaluating existing ones.
