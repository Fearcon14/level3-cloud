# Building a Cloud-Native Platform: From IaaS to PaaS

**Author:** Kevin Sinn

## Overview

**Meta-Track: Infrastructure → Platform → Product**

This track is about owning the full stack. You will get your hands on building a complete cloud platform stack end-to-end—starting from raw infrastructure, moving through Kubernetes, and ending with a production-grade PaaS product featuring APIs, UI, automation, and observability.

Every phase results in something tangible. By the end, you will have built and operated a real platform that reflects how modern cloud providers actually work.

Outcome: A fully functional PaaS on Kubernetes, running on OpenStack and SKE, complete with automation, observability, and developer tooling.

## Week 1

This week's Notion page contains all the knowledge, notes, and learnings obtained during Week 1 of the cloud-native platform journey.

📚 **[Week 1 Notion Page](https://www.notion.so/Notes-2edf7889a83b80d3bb24d64672a2613f?source=copy_link)**

### Goal

- [x] Understand and operate Infrastructure-as-a-Service by setting up and validating an OpenStack environment.

### Scope

- [x] **Understanding the OpenStack Architecture**
  - [x] How OpenStack is structured
  - [x] Responsibilities of core components:
    - [x] Virtual Machines
    - [x] Networks
    - [x] Storage and related services
- [x] **Installation and Basic Configuration** of a Single-node OpenStack cluster using DevStack
- [x] **Validation of the installation** by:
  - [x] Checking system logs
  - [x] Checking service status and health
- [x] **Deployment** of a virtual machine via OpenStack
- [x] **Architecture Analysis:**
  - [x] Analyze the current installation
  - [x] Create a technically correct architecture diagram of the setup

## Week 2

📚 **[Week 2 Notion Page](https://www.notion.so/Week-2-2f4f7889a83b800896e3e26754c9f499?source=copy_link)**

### Goal

- [x] Provision a Kubernetes cluster on OpenStack using Terraform - Infrastructure-as-Code (IaC)

### Scope

- [x] **Understanding Terraform** and Infrastructure as Code:
  - [x] What Infrastructure as Code is and what are its advantages
  - [x] Terraform basics:
    - [x] Providers
    - [x] Resources
    - [x] State
  - [x] Terraform workflow
- [x] **Provision a Virtual Machine** with Terraform using OpenStack APIs
- [x] **Installation and Configuration** of Kubernetes with a tool of your choice
- [x] **Core Kubernetes Concepts**
  - [x] Infrastructure Components: Control Plane and Worker Nodes
  - [x] Pods, Deployments, Services, Namespaces

### Bonus

- [x] **Fully Automated Installation:**
  - [x] Implement a single command to automatically provision the OpenStack Virtual Machine
  - [x] And the installation of the Kubernetes cluster

- [x] **Two-Node Kubernetes Cluster Setup:**
  - [x] Expand the installation script to provision two Virtual Machines
  - [x] And configure them to form a two-node Kubernetes cluster

## Week 3

📚 **[Week 3 Notion Page](https://www.notion.so/Week-3-2fbf7889a83b800196acd55504b150d9?source=copy_link)**

### Goal

- [x] Design and implement a Platform-as-a-Service offering on top of Kubernetes.

### Scope

- [x] **SKE Cluster Creation:** Using the STACKIT Terraform Provider to provision an SKE (STACKIT Kubernetes Engine) Cluster
- [x] **PaaS Product Implementation (e.g. Managed Database):** Design and technical implementation of a simple PaaS service.
  - [x] **Operator deployment:** Provisioning of an Operator
  - [x] **Product Component Management:** Utilization of Custom Kubernetes Resources (CRs) for the provisioning and management of product components
  - [x] **Connectivity:** Documentation and demonstration of connecting to and using the PaaS product
- [x] **Understanding Kubernetes Concepts:** Deepening knowledge of Custom Resource Definitions (CRDs) and the functioning of Operators (Reconciler Pattern)

### Bonus

- [x] **Automating the Deployment:** Introduction of a GitOps approach and CI/CD integration for automated provisioning of the SKE and the PaaS service

## Week 4

📚 **[Week 4 Notion Page](https://www.notion.so/Week-4-2fef7889a83b80b9b7bfd42b9cfbdfc8?source=copy_link)**

### Goal

- [x] Expose the PaaS product through a clean, production-ready RESTful API to enable automated provisioning and seamless integration.

### Scope

- [x] **API Development for Product Instances**: Development of a RESTful API that provides the following functions: creation, deletion, and listing of PaaS product instances, as well as retrieval of the associated connection and access data.
    - [x] API Specification: The functionality of the API must be documented in the form of OpenAPI specifications
- [x] **Unit Tests:** Implementation of simple Unit Tests for each endpoint of the developed API
- [x] **Docker Container Image** for the API creation, upload to the STACKIT Container Registry, and provisioning of the API via SKE
- [x] **Understanding the Creation Process (Create Flow):**
    - [x] * **Flowchart for the creation of a product instance:** Visualization of the individual steps
    - [x] * **Basic understanding of how a RESTful API works**

### Bonus

- [x] **Automated API Deployment**: Integration of the deployment of the RESTful API into the existing GitOps infrastructure
- [x] **Auto-Scaling and Performance Tests**:
    - [x] * **Horizontal Pod Autoscaler (HPA):** Configuration of the Kubernetes HPA for automatic scaling of the Control Plane RESTful API
    - [x] * **Performance Tests**: Implementation of performance tests for the RESTful API to verify the functionality of the HPA
- [ ] **Update Functionality**: Implementation and testing of an API endpoint that allows the updating of access data and details of the product instance


## Week 5

📚 **[Week 5 Notion Page](https://www.notion.so/Week-5-309f7889a83b803aba38c347cbcfa63b?source=copy_link)**

### Goal

Build a user-facing interface and expose the PaaS product securely via the web, ensuring a smooth and accessible developer experience.

### Scope

- [x] **Development of a user-friendly Web UI** with Vue.js or similar for interaction with the PaaS product, covering all relevant functions of the API
- [x] **Implementation of secure communication** between UI and backend APIs using JWT (JSON Web Tokens) or similar
- [x] **Deployment of Ingress controller** on SKE cluster
- [x] **Publication of the API and Web UI** on SKE cluster with SSL and free STACKIT subdomain URL
- [x] **Adaptation of the architecture diagram** to illustrate the traffic flow with API, Web UI, and Ingress

### Bonus

- [x] **Automated Deployment** of the Web UI on the existing SKE via GitOps integration
- [x] **Implementation of E2E tests** (e.g., with Cypress or Playwright) for the Web UI and the RESTful API

## Week 6 –  Implementation of Observability and Audit Logging

📚 **[Week 6 Notion Page](https://www.notion.so/Week-6-310f7889a83b80869a04de3601039686?source=copy_link)**

### Goal

Enable production-grade operations through observability and audit logging.

### Scope

- [ ] **Internal Monitoring**: Persistent telemetry data that the product operator can use for monitoring
    - [ ] * Integration of Prometheus and Grafana for monitoring the health of Kubernetes clusters and application performance
    - [ ] * Setup of Loki for collecting and analyzing application logs

- [ ] **User-Centric Monitoring:** Persistently stored logs that users can retrieve for their instances via the UI (or API)
    - [ ] * Implementation of an Audit Logging system for recording user actions (creation/modification/deletion of an instance or access data or similar)
    - [ ] * Implementation of Service Logs with relevant information for the user for security and compliance (asynchronous status changes or similar information that might be of interest to the user)

### Bonus

- [ ] **Development of a Golang SDK for the PaaS Product**

- [ ] * Creation of a Golang SDK that simplifies interaction with the PaaS product's API
- [ ] * Implementation of authentication and authorization mechanisms within the SDK
- [ ] * Provision of clear documentation and examples for using the SDK

## Final Result

By the end of this track, you will have:

* **Designed and operated a complete infrastructure stack** — from raw IaaS with OpenStack to a functional PaaS running on Kubernetes
* **Provisioned and managed Kubernetes clusters** using Infrastructure as Code with Terraform
* **Built and deployed a Kubernetes-native PaaS product**, including CRDs and Operators
* **Exposed the platform through a production-ready RESTful API** for product lifecycle management
* **Developed a secure, user-facing Web UI** that interfaces with the API and runs on Kubernetes with proper ingress and SSL
* **Implemented observability and audit logging**, enabling platform-level monitoring, logging, and traceability for both operators and users


