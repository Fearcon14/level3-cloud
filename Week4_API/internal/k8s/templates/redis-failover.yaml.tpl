apiVersion: databases.spotahome.com/v1
kind: RedisFailover
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
spec:
  sentinel:
    replicas: {{ or .SentinelReplicas 3 }}
  redis:
    replicas: {{ or .RedisReplicas 3 }}
    resources:
      requests:
        cpu: {{ or .CPURequest "100m" }}
        memory: {{ or .MemoryRequest "128Mi" }}
      limits:
        cpu: {{ or .CPULimit "500m" }}
        memory: {{ or .MemoryLimit "512Mi" }}
    storage:
      keepAfterDeletion: true
      persistentVolumeClaim:
        metadata:
          name: {{ .Name }}-data
        spec:
          storageClassName: {{ or .StorageClass "premium-perf1-stackit" }}
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: {{ or .StorageSize "1Gi" }}
