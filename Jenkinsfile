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

                    # Set Go environment
                    export PATH=$PATH:/var/lib/jenkins/go/bin
                    export GOROOT=/var/lib/jenkins/go
                    export GOPATH=/var/lib/jenkins/go-workspace

                    # Verify Go installation
                    go version

                    echo "Scanning Go code for security vulnerabilities..."

                    # Basic security pattern checks in Go files
                    echo "Checking for potential security issues..."

                    # Look for SQL injection patterns
                    echo "=== SQL Injection Patterns ==="
                    find . -name "*.go" -exec grep -Hn "fmt.Sprintf.*SELECT\\|fmt.Sprintf.*INSERT\\|fmt.Sprintf.*UPDATE\\|fmt.Sprintf.*DELETE" {} \\; > sql-injection-patterns.txt || true
                    find . -name "*.go" -exec grep -Hn "Query(.*fmt.Sprintf\\|Exec(.*fmt.Sprintf" {} \\; >> sql-injection-patterns.txt || true
                    cat sql-injection-patterns.txt || echo "No SQL injection patterns found"

                    # Look for command injection patterns
                    echo "=== Command Injection Patterns ==="
                    find . -name "*.go" -exec grep -Hn "exec.Command\\|os/exec\\|syscall.Exec" {} \\; > command-injection-patterns.txt || true
                    cat command-injection-patterns.txt || echo "No command injection patterns found"

                    # Look for potential XSS/HTML injection
                    echo "=== XSS/HTML Injection Patterns ==="
                    find . -name "*.go" -exec grep -Hn "html/template\\|text/template\\|fmt.Fprintf.*%s.*html" {} \\; > xss-patterns.txt || true
                    cat xss-patterns.txt || echo "No XSS patterns found"

                    # Look for hardcoded secrets/credentials
                    echo "=== Hardcoded Secrets Patterns ==="
                    find . -name "*.go" -exec grep -Hn "password.*=\\|secret.*=\\|token.*=\\|key.*=" {} \\; > secrets-patterns.txt || true
                    cat secrets-patterns.txt || echo "No hardcoded secrets found"

                    # Look for IDOR patterns (missing authorization checks)
                    echo "=== IDOR Patterns ==="
                    find . -name "*.go" -exec grep -Hn "GET.*/:id\\|POST.*/:id\\|PUT.*/:id\\|DELETE.*/:id" {} \\; > idor-patterns.txt || true
                    cat idor-patterns.txt || echo "No IDOR patterns found"

                    # Count vulnerabilities
                    SQL_COUNT=$(wc -l < sql-injection-patterns.txt)
                    CMD_COUNT=$(wc -l < command-injection-patterns.txt)
                    XSS_COUNT=$(wc -l < xss-patterns.txt)
                    SECRET_COUNT=$(wc -l < secrets-patterns.txt)
                    IDOR_COUNT=$(wc -l < idor-patterns.txt)

                    echo ""
                    echo "=== SECURITY SCAN SUMMARY ==="
                    echo "🔍 Potential SQL Injection patterns: $SQL_COUNT"
                    echo "🔍 Potential Command Injection patterns: $CMD_COUNT"
                    echo "🔍 Potential XSS patterns: $XSS_COUNT"
                    echo "🔍 Potential Hardcoded Secrets: $SECRET_COUNT"
                    echo "🔍 Potential IDOR patterns: $IDOR_COUNT"

                    # Check dependencies
                    if [ -f go.mod ]; then
                        echo "=== Dependency Information ==="
                        go list -json -m all > go-modules.json || true
                        echo "Dependencies exported to go-modules.json"
                    fi

                    echo "✅ SAST scan completed"
                '''
                archiveArtifacts artifacts: '*-patterns.txt, go-modules.json', allowEmptyArchive: true
            }
        }

        stage('Build Go Binary') {
            steps {
                sh '''
                    echo "🔨 Building Go application..."

                    # Set Go environment
                    export PATH=$PATH:/var/lib/jenkins/go/bin
                    export GOROOT=/var/lib/jenkins/go
                    export GOPATH=/var/lib/jenkins/go-workspace

                    # Debug: Check current directory structure
                    echo "Current directory: $(pwd)"
                    echo "Repository structure:"
                    ls -la
                    echo "Checking cmd/app directory:"
                    ls -la cmd/app/ || echo "cmd/app not found"

                    # Build from cmd/app directory
                    if [ -f cmd/app/main.go ]; then
                        echo "Found main.go in cmd/app/"

                        # Tidy dependencies from root (where go.mod should be)
                        go mod tidy

                        # Build the application pointing to cmd/app
                        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/app

                        echo "✅ Go binary built successfully from cmd/app"
                        ls -la main
                    else
                        echo "❌ main.go not found in cmd/app/"
                        echo "Available files in cmd/app:"
                        ls -la cmd/app/ || echo "cmd/app directory does not exist"
                        exit 1
                    fi
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

        stage('Deploy Infrastructure') {
            steps {
                sh '''
                    echo "🚀 Deploying infrastructure..."

                    # Create namespace
                    kubectl apply -f kubernetes/namespace.yaml

                    # Deploy database and redis
                    kubectl apply -f kubernetes/postgres-deployment.yaml
                    kubectl apply -f kubernetes/redis-deployment.yaml

                    # Wait for database to be ready
                    kubectl wait --for=condition=available --timeout=300s deployment/postgres -n auditorium || true
                    kubectl wait --for=condition=available --timeout=300s deployment/redis -n auditorium || true

                    echo "✅ Infrastructure deployed successfully"
                '''
            }
        }

        stage('Database Migration & Seeding') {
            steps {
                sh '''
                    echo "📋 Setting up migration and seeder ConfigMaps..."

                    # Create or update migration ConfigMap from actual files
                    kubectl create configmap migration-files \
                        --from-file=database/migration/ \
                        --namespace=auditorium \
                        --dry-run=client -o yaml | kubectl apply -f -

                    # Create or update seeder ConfigMap from actual files
                    kubectl create configmap seeder-files \
                        --from-file=database/seeder/ \
                        --namespace=auditorium \
                        --dry-run=client -o yaml | kubectl apply -f -

                    echo "✅ ConfigMaps created/updated from actual migration and seeder files"

                    # Wait for database to be ready
                    echo "⏳ Waiting for database to be ready..."
                    kubectl wait --for=condition=available --timeout=300s deployment/postgres -n auditorium

                    echo "🗃️ Running database migration..."

                    # Create and run migration job
                    kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migration-${BUILD_NUMBER}
  namespace: auditorium
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: migrate
          image: migrate/migrate:v4.17.0
          command: ["/migrate"]
          args:
            - "-path=/migrations"
            - "-database=postgres://auditorium-reservation-backend:thisisasamplepassword@postgres-service:5432/auditorium-reservation-backend?sslmode=disable"
            - "up"
          volumeMounts:
            - name: migrations
              mountPath: /migrations
      volumes:
        - name: migrations
          configMap:
            name: migration-files
EOF

                    # Wait for migration to complete
                    echo "⏳ Waiting for migration to complete..."
                    kubectl wait --for=condition=complete --timeout=300s job/db-migration-${BUILD_NUMBER} -n auditorium

                    # Check migration logs
                    echo "📋 Migration logs:"
                    kubectl logs job/db-migration-${BUILD_NUMBER} -n auditorium

                    echo "🌱 Running database seeding..."

                    # Create and run seeder job with duplicate check
                    kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: db-seeder-${BUILD_NUMBER}
  namespace: auditorium
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: seeder
          image: postgres:17.2
          command: ['sh', '-c']
          args:
            - |
              echo "Checking if seeder data already exists..."

              # Check if seeder data exists
              EXISTING_USERS=\\$(psql "postgres://auditorium-reservation-backend:thisisasamplepassword@postgres-service:5432/auditorium-reservation-backend" -t -c "SELECT COUNT(*) FROM users WHERE email LIKE '%@seeder.nathakusuma.com';" 2>/dev/null | xargs)

              if [ "\\$EXISTING_USERS" -gt 0 ]; then
                echo "Seeder data already exists (\\$EXISTING_USERS users found), skipping..."
              else
                echo "No seeder data found, running seeder..."
                psql "postgres://auditorium-reservation-backend:thisisasamplepassword@postgres-service:5432/auditorium-reservation-backend" -f /seeders/dev.up.sql
                echo "✅ Seeding completed!"
              fi
          volumeMounts:
            - name: seeders
              mountPath: /seeders
      volumes:
        - name: seeders
          configMap:
            name: seeder-files
EOF

                    # Wait for seeding to complete
                    echo "⏳ Waiting for seeding to complete..."
                    kubectl wait --for=condition=complete --timeout=300s job/db-seeder-${BUILD_NUMBER} -n auditorium

                    # Check seeder logs
                    echo "🌱 Seeder logs:"
                    kubectl logs job/db-seeder-${BUILD_NUMBER} -n auditorium

                    # Verify database setup
                    echo "🔍 Verifying database setup..."
                    kubectl exec deployment/postgres -n auditorium -- psql -U auditorium-reservation-backend -d auditorium-reservation-backend -c "\\\\dt"

                    # Count records
                    echo "📊 Database record counts:"
                    kubectl exec deployment/postgres -n auditorium -- psql -U auditorium-reservation-backend -d auditorium-reservation-backend -c "SELECT 'Users: ' || COUNT(*) FROM users; SELECT 'Conferences: ' || COUNT(*) FROM conferences; SELECT 'Registrations: ' || COUNT(*) FROM registrations;"

                    # Clean up old migration/seeder jobs (keep only last 3)
                    kubectl get jobs -n auditorium | grep "db-migration-" | head -n -3 | awk '{print \\$1}' | xargs -r kubectl delete job -n auditorium || true
                    kubectl get jobs -n auditorium | grep "db-seeder-" | head -n -3 | awk '{print \\$1}' | xargs -r kubectl delete job -n auditorium || true

                    echo "✅ Database migration and seeding completed successfully!"
                '''
            }
        }

        stage('Deploy Backend Application') {
            steps {
                sh '''
                    echo "🚀 Deploying backend application..."

                    # Update backend deployment with build number
                    sed -i "s|\\${BUILD_NUMBER}|${BUILD_NUMBER}|g" kubernetes/backend-deployment.yaml

                    # Deploy backend
                    kubectl apply -f kubernetes/backend-deployment.yaml

                    # Wait for backend deployment
                    kubectl wait --for=condition=available --timeout=300s deployment/auditorium-backend -n auditorium || true

                    echo "✅ Backend application deployed successfully"
                '''
            }
        }

        stage('DAST - Dynamic Security Testing') {
            steps {
                sh '''
                    echo "🔍 Running Dynamic Application Security Testing..."

                    # Get application URL
                    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
                    APP_URL="http://$NODE_IP:30081"

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
                    echo "Testing authentication endpoints..." >> vulnerability-tests.txt
                    curl -X POST "$APP_URL/api/v1/auth/login" \
                         -H "Content-Type: application/json" \
                         -d '{"email": "admin'"'"' OR 1=1 --", "password": "test"}' \
                         -w "\nSQL Injection Test - Status: %{http_code}\n" \
                         -s >> vulnerability-tests.txt || true
                    
                    echo "Testing user access endpoints..." >> vulnerability-tests.txt
                    for id in 1 2 3 999 -1; do
                        curl "$APP_URL/api/v1/users/$id" \
                             -w "\nIDOR Test ID $id - Status: %{http_code}\n" \
                             -s >> vulnerability-tests.txt || true
                    done
                    
                    echo "Testing admin endpoints without auth..." >> vulnerability-tests.txt
                    curl "$APP_URL/api/v1/admin/users" \
                         -w "\nUnauth Admin Access - Status: %{http_code}\n" \
                         -s >> vulnerability-tests.txt || true
                    
                    echo "Testing for exposed files..." >> vulnerability-tests.txt
                    curl "$APP_URL/.env" \
                         -w "\n.env file - Status: %{http_code}\n" \
                         -s >> vulnerability-tests.txt || true
                    
                    curl "$APP_URL/config" \
                         -w "\nConfig endpoint - Status: %{http_code}\n" \
                         -s >> vulnerability-tests.txt || true
                    
                    echo "✅ Manual endpoint testing completed. Check vulnerability-tests.txt for results." >> vulnerability-tests.txt
                    # Create endpoints url
                    echo "$APP_URL/" >> url.txt
                    echo "$APP_URL/api/v1/auth/login" >> url.txt
                    echo "$APP_URL/api/v1/users/1" >> url.txt
                    echo "$APP_URL/api/v1/admin/users" >> url.txt
                    echo "$APP_URL/.env" >> url.txt
                    
                    # Run OWASP ZAP scan
                     echo "📁 Fixing workspace permission"
                    chmod -R 777 $PWD
                    
                    echo "Running OWASP ZAP baseline scan..."
                    docker pull zaproxy/zap-stable
                    docker run -v $PWD:/zap/wrk -t zaproxy/zap-stable zap-baseline.py -t $APP_URL -J zap-report.json -r zap-report.html -z "-cmd runurls /zap/wrk/url.txt" || true

                    echo "✅ Security testing completed"
                    echo "📊 Check zap-report.json for detailed vulnerability report"
                '''
                archiveArtifacts artifacts: 'zap-report.json', allowEmptyArchive: true
                archiveArtifacts artifacts: 'zap-report.html', allowEmptyArchive: true
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
                echo "🌐 Application accessible at: http://$NODE_IP:30081"
                echo "📊 Security reports available in Jenkins artifacts"
                echo ""
                echo "🗃️ Database Status:"
                echo "📋 Migration and seeding completed automatically"
                echo "📊 Check database for tables and seeded data"
                echo ""
                echo "🎯 DAST found vulnerabilities in your backend."
                echo "📋 Next steps:"
                echo "1. Review zap-report.json and security scan results"
                echo "2. Identify 3 vulnerabilities to fix"
                echo "3. Fix them one by one and re-run pipeline"
                echo "4. Database will auto-migrate on each deployment"
            '''
        }
        failure {
            echo "❌ Pipeline failed!"
            sh '''
                echo "Debugging information:"
                kubectl get pods -n auditorium || true
                kubectl describe pods -n auditorium || true
                kubectl logs -l app=auditorium-backend -n auditorium --tail=50 || true

                # Check migration/seeder job logs if they exist
                kubectl logs job/db-migration-${BUILD_NUMBER} -n auditorium || true
                kubectl logs job/db-seeder-${BUILD_NUMBER} -n auditorium || true
            '''
        }
    }
}
