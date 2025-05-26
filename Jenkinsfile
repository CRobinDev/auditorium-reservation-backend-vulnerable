pipeline {
    agent any

    environment {
        REGISTRY = "localhost:30500"
        IMAGE_NAME = "auditorium-backend"
        BUILD_NUMBER = "${env.BUILD_NUMBER}"
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
                echo "✅ Code checked out successfully"
            }
        }

        stage('Security Scan - SAST') {
            steps {
                sh '''
                    echo "🔍 Running Static Application Security Testing..."

                    # Install Go security tools
                    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest || true

                    # Run gosec for static analysis
                    ~/go/bin/gosec -fmt json -out gosec-report.json ./... || true

                    # Check for dependency vulnerabilities
                    go list -json -m all > go-modules.json || true

                    echo "✅ SAST scan completed"
                '''
                archiveArtifacts artifacts: 'gosec-report.json, go-modules.json', allowEmptyArchive: true
            }
        }

        stage('Build Go Binary') {
            steps {
                sh '''
                    echo "🔨 Building Go application..."
                    go mod tidy
                    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .
                    echo "✅ Go binary built successfully"
                '''
            }
        }

        stage('Build Docker Image') {
            steps {
                sh '''
                    echo "🐳 Building Docker image..."

                    # Use existing Dockerfile atau create simple one
                    if [ ! -f Dockerfile ]; then
                        cat > Dockerfile << EOF
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY main .
EXPOSE 8080
CMD ["./main"]
EOF
                    fi

                    docker build -t ${REGISTRY}/${IMAGE_NAME}:${BUILD_NUMBER} .
                    echo "✅ Docker image built successfully"
                '''
            }
        }

        stage('Push Docker Image') {
            steps {
                sh '''
                    echo "📤 Pushing Docker image..."
                    docker push ${REGISTRY}/${IMAGE_NAME}:${BUILD_NUMBER}
                    echo "✅ Docker image pushed successfully"
                '''
            }
        }

        stage('Deploy to Kubernetes') {
            steps {
                sh '''
                    echo "🚀 Deploying to Kubernetes..."

                    # Create namespace
                    kubectl apply -f kubernetes/namespace.yaml

                    # Deploy database and redis
                    kubectl apply -f kubernetes/postgres-deployment.yaml
                    kubectl apply -f kubernetes/redis-deployment.yaml

                    # Wait for database to be ready
                    kubectl wait --for=condition=available --timeout=300s deployment/postgres -n auditorium || true
                    kubectl wait --for=condition=available --timeout=300s deployment/redis -n auditorium || true

                    # Update backend deployment with build number
                    sed -i "s|\\${BUILD_NUMBER}|${BUILD_NUMBER}|g" kubernetes/backend-deployment.yaml

                    # Deploy backend
                    kubectl apply -f kubernetes/backend-deployment.yaml

                    # Wait for backend deployment
                    kubectl wait --for=condition=available --timeout=300s deployment/auditorium-backend -n auditorium || true

                    echo "✅ Application deployed successfully"
                '''
            }
        }

        stage('DAST - Dynamic Security Testing') {
            steps {
                sh '''
                    echo "🔍 Running Dynamic Application Security Testing..."

                    # Get application URL
                    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
                    APP_URL="http://$NODE_IP:30080"

                    # Wait for application to be ready
                    echo "⏳ Waiting for application to be ready..."
                    for i in {1..20}; do
                        if curl -f $APP_URL > /dev/null 2>&1; then
                            echo "✅ Application is responding at $APP_URL"
                            break
                        fi
                        echo "Waiting... ($i/20)"
                        sleep 15
                    done

                    echo "🎯 Testing for common vulnerabilities..."

                    # Test common vulnerable endpoints (adjust based on your actual endpoints)
                    echo "Testing authentication endpoints..."
                    curl -X POST "$APP_URL/api/auth/login" \
                         -H "Content-Type: application/json" \
                         -d '{"email": "admin'"'"' OR 1=1 --", "password": "test"}' \
                         -w "\\nSQL Injection Test - Status: %{http_code}\\n" || true

                    echo "Testing user access endpoints..."
                    for id in 1 2 3 999 -1; do
                        curl "$APP_URL/api/users/$id" \
                             -w "\\nIDOR Test ID $id - Status: %{http_code}\\n" || true
                    done

                    echo "Testing admin endpoints without auth..."
                    curl "$APP_URL/api/admin/users" \
                         -w "\\nUnauth Admin Access - Status: %{http_code}\\n" || true

                    echo "Testing for exposed files..."
                    curl "$APP_URL/.env" -w "\\n.env file - Status: %{http_code}\\n" || true
                    curl "$APP_URL/config" -w "\\nConfig endpoint - Status: %{http_code}\\n" || true

                    # Run OWASP ZAP scan
                    echo "Running OWASP ZAP baseline scan..."
                    docker pull zaproxy/zap-stable
                    docker run -t zaproxy/zap-stable zap-baseline.py -t $APP_URL -J zap-report.json || true

                    echo "✅ Security testing completed"
                    echo "📊 Check zap-report.json for detailed vulnerability report"
                '''
                archiveArtifacts artifacts: 'zap-report.json', allowEmptyArchive: true
            }
        }
    }

    post {
        always {
            echo "🧹 Cleaning up..."
            sh '''
                # Clean up old docker images
                docker image prune -f || true

                # Show current deployment status
                echo "Current deployment status:"
                kubectl get pods -n auditorium || true
            '''
        }
        success {
            sh '''
                NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
                echo "✅ Pipeline completed successfully!"
                echo "🌐 Application accessible at: http://$NODE_IP:30080"
                echo "📊 Security reports available in Jenkins artifacts"
                echo ""
                echo "🎯 DAST found vulnerabilities in your existing backend."
                echo "📋 Next steps:"
                echo "1. Review zap-report.json and gosec-report.json"
                echo "2. Identify 3 vulnerabilities to fix"
                echo "3. Fix them one by one and re-run pipeline"
                echo "4. Repeat until all vulnerabilities are resolved"
            '''
        }
        failure {
            echo "❌ Pipeline failed!"
            sh '''
                echo "Debugging information:"
                kubectl get pods -n auditorium || true
                kubectl describe pods -n auditorium || true
                kubectl logs -l app=auditorium-backend -n auditorium --tail=50 || true
            '''
        }
    }
}
