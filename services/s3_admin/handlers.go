package s3_admin

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strconv"
)

type createBucketRequest struct {
	Name string `json:"name"`
}

func ListEndpointsHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpoints, err := service.ListEndpoints(r.Context())
		if err != nil {
			slog.Error("failed to list s3 endpoints", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, endpoints)
	}
}

func BucketsHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpointName := r.PathValue("endpoint")

		switch r.Method {
		case http.MethodGet:
			buckets, err := service.ListBuckets(r.Context(), endpointName)
			if err != nil {
				slog.Error("failed to list s3 buckets", slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			writeJSON(w, http.StatusOK, buckets)
		case http.MethodPost:
			var req createBucketRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			if BucketNameFromURLValue(req.Name) == "" {
				http.Error(w, "bucket name is required", http.StatusBadRequest)
				return
			}

			if err := service.CreateBucket(r.Context(), endpointName, req.Name); err != nil {
				slog.Error("failed to create s3 bucket", slog.String("bucket", req.Name), slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func BucketByNameHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpointName := r.PathValue("endpoint")
		bucketName := BucketNameFromURLValue(r.PathValue("bucket"))

		if bucketName == "" {
			http.Error(w, "bucket name is required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodDelete:
			if err := service.DeleteBucket(r.Context(), endpointName, bucketName); err != nil {
				slog.Error("failed to delete s3 bucket", slog.String("bucket", bucketName), slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func ObjectsHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpointName := r.PathValue("endpoint")
		bucketName := BucketNameFromURLValue(r.PathValue("bucket"))
		if bucketName == "" {
			http.Error(w, "bucket name is required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			prefix := r.URL.Query().Get("prefix")
			objects, err := service.ListObjects(r.Context(), endpointName, bucketName, prefix)
			if err != nil {
				slog.Error("failed to list s3 objects", slog.String("bucket", bucketName), slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, objects)
		case http.MethodPost:
			if err := handleUploadObject(w, r, service, endpointName, bucketName); err != nil {
				slog.Error("failed to upload s3 object", slog.String("bucket", bucketName), slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			objectKey := ObjectKeyFromURLValue(r.URL.Query().Get("key"))
			if objectKey == "" {
				http.Error(w, "object key is required", http.StatusBadRequest)
				return
			}

			if err := service.DeleteObject(r.Context(), endpointName, bucketName, objectKey); err != nil {
				slog.Error("failed to delete s3 object", slog.String("bucket", bucketName), slog.String("key", objectKey), slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func DownloadObjectHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpointName := r.PathValue("endpoint")
		bucketName := BucketNameFromURLValue(r.PathValue("bucket"))
		objectKey := ObjectKeyFromURLValue(r.URL.Query().Get("key"))
		if bucketName == "" || objectKey == "" {
			http.Error(w, "bucket and object key are required", http.StatusBadRequest)
			return
		}

		object, info, err := service.DownloadObject(r.Context(), endpointName, bucketName, objectKey)
		if err != nil {
			slog.Error("failed to download s3 object", slog.String("bucket", bucketName), slog.String("key", objectKey), slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer object.Close()

		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, path.Base(objectKey)))
		if info.ContentType != "" {
			w.Header().Set("Content-Type", info.ContentType)
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		if info.Size >= 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(info.Size, 10))
		}

		if _, err := io.Copy(w, object); err != nil {
			slog.Error("failed streaming s3 object", slog.String("bucket", bucketName), slog.String("key", objectKey), slog.String("error", err.Error()))
		}
	}
}

func handleUploadObject(w http.ResponseWriter, r *http.Request, service *Service, endpointName string, bucketName string) error {
	if err := r.ParseMultipartForm(128 << 20); err != nil {
		return fmt.Errorf("failed to parse multipart upload: %w", err)
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		return fmt.Errorf("missing upload file: %w", err)
	}
	defer file.Close()

	objectKey := SafeObjectKey(r.FormValue("objectKey"))
	if objectKey == "" {
		objectKey = SafeObjectKey(fileHeader.Filename)
	}
	if objectKey == "" {
		return fmt.Errorf("object key is required")
	}

	contentType := fileHeader.Header.Get("Content-Type")
	return service.UploadObject(r.Context(), endpointName, bucketName, objectKey, file, fileHeader.Size, contentType)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
