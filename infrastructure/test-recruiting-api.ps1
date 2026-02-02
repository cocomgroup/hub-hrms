#!/usr/bin/env pwsh
# Test Hub-HRMS Recruiting GraphQL API
# Run this after starting the backend server

$ErrorActionPreference = "Stop"
$API_URL = "http://localhost:8080/graphql"

function Invoke-GraphQL {
    param(
        [string]$Query,
        [hashtable]$Variables = @{}
    )
    
    $body = @{
        query = $Query
    }
    
    if ($Variables.Count -gt 0) {
        $body.variables = $Variables
    }
    
    try {
        $response = Invoke-RestMethod -Uri $API_URL -Method Post -Body ($body | ConvertTo-Json -Depth 10) -ContentType "application/json"
        return $response
    } catch {
        Write-Host "Error: $_" -ForegroundColor Red
        return $null
    }
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Hub-HRMS Recruiting API Tests" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Check if server is running
Write-Host "Test 1: Checking if GraphQL server is running..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/api/health" -Method Get -TimeoutSec 5
    Write-Host "[PASS] Server is running" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "[FAIL] Server is not responding. Please start the backend server first:" -ForegroundColor Red
    Write-Host "  cd C:\Users\evanh\projects\cocom\hub-hrms\backend" -ForegroundColor Yellow
    Write-Host "  go run .\cmd\server\main.go" -ForegroundColor Yellow
    exit 1
}

# Test 2: Fetch all jobs
Write-Host "Test 2: Fetching all published jobs..." -ForegroundColor Yellow
$jobsQuery = @"
query GetJobs {
  jobs(filters: { status: "PUBLISHED" }, limit: 5) {
    id
    title
    department
    location
    employmentType
    applicationCount
    viewCount
    remoteWork
  }
}
"@

$jobsResult = Invoke-GraphQL -Query $jobsQuery
if ($jobsResult.data.jobs) {
    Write-Host "[PASS] Found $($jobsResult.data.jobs.Count) jobs" -ForegroundColor Green
    foreach ($job in $jobsResult.data.jobs) {
        Write-Host "  - [$($job.id)] $($job.title) ($($job.department))" -ForegroundColor White
        Write-Host "    Location: $($job.location) | Type: $($job.employmentType)" -ForegroundColor Gray
        Write-Host "    Applications: $($job.applicationCount) | Views: $($job.viewCount)" -ForegroundColor Gray
    }
    Write-Host ""
} else {
    Write-Host "[FAIL] No jobs found or query failed" -ForegroundColor Red
    Write-Host ""
}

# Test 3: Get job details
if ($jobsResult.data.jobs -and $jobsResult.data.jobs.Count -gt 0) {
    $firstJobId = $jobsResult.data.jobs[0].id
    Write-Host "Test 3: Getting details for job ID: $firstJobId..." -ForegroundColor Yellow
    
    $jobDetailsQuery = @"
query GetJobDetails(`$jobId: ID!) {
  job(id: `$jobId) {
    id
    title
    department
    location
    description
    requirements
    salaryRange {
      min
      max
      currency
    }
    skills
    applicationCount
    viewCount
  }
}
"@
    
    $jobDetailsResult = Invoke-GraphQL -Query $jobDetailsQuery -Variables @{ jobId = $firstJobId }
    if ($jobDetailsResult.data.job) {
        $job = $jobDetailsResult.data.job
        Write-Host "[PASS] Job Details Retrieved" -ForegroundColor Green
        Write-Host "  Title: $($job.title)" -ForegroundColor White
        Write-Host "  Department: $($job.department)" -ForegroundColor White
        Write-Host "  Location: $($job.location)" -ForegroundColor White
        if ($job.salaryRange) {
            Write-Host "  Salary: `$$($job.salaryRange.min) - `$$($job.salaryRange.max) $($job.salaryRange.currency)" -ForegroundColor White
        }
        if ($job.skills) {
            Write-Host "  Skills: $($job.skills -join ', ')" -ForegroundColor White
        }
        Write-Host ""
    }
}

# Test 4: Increment job view count
if ($jobsResult.data.jobs -and $jobsResult.data.jobs.Count -gt 0) {
    $firstJobId = $jobsResult.data.jobs[0].id
    Write-Host "Test 4: Incrementing view count for job $firstJobId..." -ForegroundColor Yellow
    
    $incrementViewQuery = @"
mutation IncrementView(`$jobId: ID!) {
  incrementJobView(id: `$jobId) {
    id
    viewCount
  }
}
"@
    
    $incrementResult = Invoke-GraphQL -Query $incrementViewQuery -Variables @{ jobId = $firstJobId }
    if ($incrementResult.data.incrementJobView) {
        Write-Host "[PASS] View count incremented to: $($incrementResult.data.incrementJobView.viewCount)" -ForegroundColor Green
        Write-Host ""
    }
}

# Test 5: Submit a test application
if ($jobsResult.data.jobs -and $jobsResult.data.jobs.Count -gt 0) {
    $firstJobId = $jobsResult.data.jobs[0].id
    Write-Host "Test 5: Submitting a test application to job $firstJobId..." -ForegroundColor Yellow
    
    $submitAppQuery = @"
mutation SubmitApplication(`$input: ApplicationInput!) {
  submitApplication(input: `$input) {
    id
    status
    appliedDate
    candidate {
      id
      firstName
      lastName
      email
    }
    job {
      title
    }
    aiScore {
      overall
      recommendation
    }
  }
}
"@
    
    $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
    $appInput = @{
        jobId = $firstJobId
        firstName = "Test"
        lastName = "Candidate"
        email = "test.candidate.$timestamp@example.com"
        phone = "+1-555-0100"
        resumeUrl = "https://example.com/resume.pdf"
        coverLetter = "This is a test application submitted via PowerShell script."
        yearsOfExperience = 5
        currentLocation = "San Francisco, CA"
        willingToRelocate = $true
        expectedSalary = 120000.0
        availability = "2 weeks"
    }
    
    $submitResult = Invoke-GraphQL -Query $submitAppQuery -Variables @{ input = $appInput }
    if ($submitResult.data.submitApplication) {
        $app = $submitResult.data.submitApplication
        Write-Host "[PASS] Application submitted successfully!" -ForegroundColor Green
        Write-Host "  Application ID: $($app.id)" -ForegroundColor White
        Write-Host "  Candidate: $($app.candidate.firstName) $($app.candidate.lastName)" -ForegroundColor White
        Write-Host "  Email: $($app.candidate.email)" -ForegroundColor White
        Write-Host "  Job: $($app.job.title)" -ForegroundColor White
        Write-Host "  Status: $($app.status)" -ForegroundColor White
        Write-Host "  Applied: $($app.appliedDate)" -ForegroundColor White
        if ($app.aiScore) {
            Write-Host "  AI Score: $($app.aiScore.overall)/10" -ForegroundColor White
            Write-Host "  Recommendation: $($app.aiScore.recommendation)" -ForegroundColor White
        }
        Write-Host ""
        
        # Store application ID for next test
        $script:testAppId = $app.id
    } else {
        Write-Host "[FAIL] Failed to submit application" -ForegroundColor Red
        if ($submitResult.errors) {
            Write-Host "  Errors: $($submitResult.errors | ConvertTo-Json)" -ForegroundColor Red
        }
        Write-Host ""
    }
}

# Test 6: Retrieve the submitted application
if ($script:testAppId) {
    Write-Host "Test 6: Retrieving application details..." -ForegroundColor Yellow
    
    $getAppQuery = @"
query GetApplication(`$appId: ID!) {
  application(id: `$appId) {
    id
    status
    appliedDate
    candidate {
      firstName
      lastName
      email
      phone
    }
    job {
      title
      department
    }
    coverLetter
    yearsOfExperience
    expectedSalary
    aiScore {
      overall
      insights
      strengths
      concerns
      recommendation
    }
  }
}
"@
    
    $appResult = Invoke-GraphQL -Query $getAppQuery -Variables @{ appId = $script:testAppId }
    if ($appResult.data.application) {
        $app = $appResult.data.application
        Write-Host "[PASS] Application retrieved successfully" -ForegroundColor Green
        Write-Host "  Candidate: $($app.candidate.firstName) $($app.candidate.lastName)" -ForegroundColor White
        Write-Host "  Phone: $($app.candidate.phone)" -ForegroundColor White
        Write-Host "  Years of Experience: $($app.yearsOfExperience)" -ForegroundColor White
        Write-Host "  Expected Salary: `$$($app.expectedSalary)" -ForegroundColor White
        if ($app.aiScore) {
            Write-Host "  AI Analysis:" -ForegroundColor Cyan
            Write-Host "    Score: $($app.aiScore.overall)/10" -ForegroundColor White
            Write-Host "    Insights: $($app.aiScore.insights)" -ForegroundColor White
            if ($app.aiScore.strengths) {
                Write-Host "    Strengths: $($app.aiScore.strengths -join ', ')" -ForegroundColor White
            }
            if ($app.aiScore.concerns) {
                Write-Host "    Concerns: $($app.aiScore.concerns -join ', ')" -ForegroundColor White
            }
        }
        Write-Host ""
    }
}

# Test 7: Filter jobs by criteria
Write-Host "Test 7: Testing job filters..." -ForegroundColor Yellow

$filterQuery = @"
query FilterJobs {
  jobs(filters: { 
    remoteWork: true
    employmentType: "FULL_TIME"
  }, limit: 5) {
    id
    title
    department
    location
    remoteWork
    employmentType
  }
}
"@

$filterResult = Invoke-GraphQL -Query $filterQuery
if ($filterResult.data.jobs) {
    Write-Host "[PASS] Found $($filterResult.data.jobs.Count) remote full-time jobs" -ForegroundColor Green
    foreach ($job in $filterResult.data.jobs) {
        Write-Host "  - $($job.title) at $($job.department)" -ForegroundColor White
    }
    Write-Host ""
}

# Summary
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "All tests completed!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Open GraphQL Playground: http://localhost:8080/graphql" -ForegroundColor White
Write-Host "  2. Test from your Svelte frontend (hr-recruiting)" -ForegroundColor White
Write-Host "  3. Review the full testing guide: RECRUITING_API_TESTING_GUIDE.md" -ForegroundColor White
Write-Host ""