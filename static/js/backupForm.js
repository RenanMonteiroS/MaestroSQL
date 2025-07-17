let currentStep = 1;
const totalSteps = 4;
let authModal;
let abortController = new AbortController()

//When the document is loaded...
document.addEventListener('DOMContentLoaded', async function() {   
    if (window.appConfig.authenticationMethods.length > 0) {
        const isAuthResult = await isAuth();
        authModal = new bootstrap.Modal(document.getElementById('authModal'), {
                    backdrop: 'static',
                    keyboard: false
        });
        
        if (!isAuthResult) {
            authModal.show()
        } else {
            updateAuthStatus(true, isAuthResult.data.user);
        }

        
    }
});



/**
 * Handles the modal for perform login, from the fade in/out and alerts to HTTP request to the authenticator.
 * @throws {Error} Throws an error then the backend returns a bad HTTP status code
*/
async function modalLogin() {
    const loginBtn = document.getElementById('modalLoginBtn');
    const loginText = loginBtn.querySelector('.login-text');
    const loadingSpinner = loginBtn.querySelector('.loading-spinner');
    
    try {
        // Shows loading
        loginText.style.display = 'none';
        loadingSpinner.style.display = 'inline';
        loginBtn.disabled = true;

        const userDataPayload = {
            "email": document.getElementById('modalEmail').value,
            "password": document.getElementById('modalPassword').value,
            "mfaKey": document.getElementById('modalMfaToken').value,
        };

        headers = getHeaders()
        const response = await fetch(`/login?method=osi`, {
            method: 'POST',
            headers: getHeaders(),
            body: JSON.stringify(userDataPayload)
        });

        const result = {"status": response.status, "body": await response.json()};

        if (!response.ok) {
            throw new Error(`${result.body.errors.osiMsg || result.body.errors.request}`);
        }

        document.getElementById('modalMfaToken').value = "";
        document.getElementById('modalPassword').value = "";
        document.getElementById('modalEmail').value = "";
        
        // Closes the modal
        authModal.hide();

        // Shows a success message
        setTimeout(() => {
            alert(window.appConfig.translations.userAuthenticatedSuccess.replace("{email}", result.body.data.user));
        }, 300);

        // Updates the authentication status in the auth card
        updateAuthStatus(true, result.body.data.user);

    } catch (err) {
        console.error('Authentication error:', err.message);
        alert(window.appConfig.translations.authenticationError.replace("{errorMessage}", err.message));
    } finally {
        // Restores the button
        loginText.style.display = 'inline';
        loadingSpinner.style.display = 'none';
        loginBtn.disabled = false;
    }
}

/**
 * Updates the authentication status, present in the auth card.
 * @param {boolean} authenticated - If the user is authenticated or not
 * @param {string} - The e-mail of the user who is authenticated
*/
function updateAuthStatus(authenticated, email = '') {
    const authStatus = document.getElementById('auth-status');
    const authBtn = document.getElementById('auth-btn');
    const logoutBtn = document.getElementById('logout-btn');
    
    if (authStatus) {
        if (authenticated) {
            authStatus.className = 'auth-status authenticated mb-3';
            authStatus.innerHTML = `
                <i class="fas fa-check-circle"></i>
                <span>${window.appConfig.translations.authenticatedAs.replace("{email}", decodeURIComponent(email))}</span>
            `;
            if (authBtn) authBtn.style.display = 'none';
            if (logoutBtn) logoutBtn.style.display = 'block';
        } else {
            authStatus.className = 'auth-status unauthenticated mb-3';
            authStatus.innerHTML = `
                <i class="fas fa-exclamation-circle"></i>
                <span>${window.appConfig.translations.notAuthenticated}</span>
            `;
            if (authBtn) authBtn.style.display = 'block';
            if (logoutBtn) logoutBtn.style.display = 'none';
        }
    }
}

/**
 * Updates the navbar button colors
*/
function updateStepNavigation() {
    for (let i = 1; i <= totalSteps; i++) {
        const stepNav = document.getElementById(`step-nav-${i}`);
        const stepContent = document.getElementById(`step-${i}`);
        
        if (i < currentStep) {
            stepNav.className = 'step-item completed';
            stepContent.classList.remove('active');
        } else if (i === currentStep) {
            stepNav.className = 'step-item active';
            stepContent.classList.add('active');
        } else {
            stepNav.className = 'step-item inactive';
            stepContent.classList.remove('active');
        }
    }

    // Updates the navbar buttons
    document.getElementById('prev-btn').classList.toggle('d-none', currentStep === 1);
    
    if (currentStep === totalSteps) {
        document.getElementById('next-btn').classList.add('d-none');
        document.getElementById('execute-btn').classList.remove('d-none');
        generateSummary();
    } else {
        document.getElementById('next-btn').classList.remove('d-none');
        document.getElementById('execute-btn').classList.add('d-none');
    }
}

/**
 * Once the "next" button is pressed, this function is called. It checks the current step and manage the form workflow
*/
async function nextStep() {
    if (!validateCurrentStep()) {
        return;
    }

    // Makes a POST request to /connect, establishing a connection with the database
    if (currentStep === 1) {
        const connectBtn = document.getElementById('next-btn');
        const originalText = connectBtn.innerHTML;
        
        try {
            connectBtn.innerHTML = `<i class="fas fa-spinner fa-spin me-1"></i> ${window.appConfig.translations.connecting}`;
            connectBtn.disabled = true;

            const connectionData = {
                host: document.getElementById('host').value,
                port: document.getElementById('port').value,
                user: document.getElementById('user').value,
                password: document.getElementById('password').value,
                instance: document.getElementById('instance').value,
                encryption: document.getElementById('encryption').value,
                trustServerCertificate: document.getElementById('trustServerCertificate').checked ? true : false
            };

            const response = await fetch(`/connect`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify(connectionData)
            });

            const result = {"status": response.status, "body": await response.json()};

            if (!response.ok) {
                if (response.status == 401) {
                    authModal.show();
                    throw new Error(`${result.body.message}`);
                }
                throw new Error(`${result.body.errors.connect}`);
            }

            console.log(`Connection established: ${result.body.message}. Server: ${result.body.data.server}`);
            
        } catch (error) {
            console.error('Error connecting:', error.message);
            alert(window.appConfig.translations.connectionError.replace("{errorMessage}", error.message));
            connectBtn.innerHTML = originalText;
            connectBtn.disabled = false;
            return;
        } finally {
            connectBtn.innerHTML = originalText;
            connectBtn.disabled = false;
        }
    }

    if (currentStep < totalSteps) {
        currentStep++;
        updateStepNavigation();
        
        // Loads the databases allocated in the server
        if (currentStep === 3 && document.querySelector('input[name="operation"]:checked').value === 'backup') {
            const databaseContainer = document.querySelector('#step-3 .row');
            const databaseSelectButtons = document.querySelector("#databaseSelectButtons");
            const backupPath = document.querySelector("#backup-path");

            databaseContainer.innerHTML = '';
            databaseSelectButtons.innerHTML = '';
            backupPath.innerHTML = 
                `<div class="col-md-12 mb-3">
                        <label for="path" class="form-label">${window.appConfig.translations.backupsPath}</label>
                        <i class="fas fa-info-circle info-icon" 
                                data-bs-toggle="tooltip" 
                                data-bs-placement="right" 
                                title="${window.appConfig.translations.backupsPathTooltipBackup}">
                        </i>
                        <input type="text" class="form-control" id="path" placeholder="C:/Backups">
                </div>`;
            
            await loadDatabases();
        } else if (currentStep === 3 && document.querySelector('input[name="operation"]:checked').value === 'restore') {
            const databaseContainer = document.querySelector('#step-3 .row');
            const databaseSelectButtons = document.querySelector("#databaseSelectButtons");
            const backupPath = document.querySelector("#backup-path");

            databaseContainer.innerHTML = '';
            databaseSelectButtons.innerHTML = '';
            backupPath.innerHTML = 
                 `<div class="col-md-12 mb-3">
                        <label for="path" class="form-label">${window.appConfig.translations.backupsPath}</label>
                        <i class="fas fa-info-circle info-icon" 
                                data-bs-toggle="tooltip" 
                                data-bs-placement="right" 
                                title=${window.appConfig.translations.backupsPathTooltipRestore}">
                        </i>
                        <div class="input-group">
                            <input type="text" class="form-control" id="path" placeholder="C:/Backups/">
                            <button class="btn btn-outline-secondary" type="button" id="list-backups-btn" onclick="listBackups()">${window.appConfig.translations.listBackups}</button>
                        </div>
                </div>`;
        }
    }
}

/**
 * Makes a POST request to the backend, collecting all the .bak files in the given path and generating the dynamic HTML content
 * @throws {Error} Throws an error then the backend returns a bad HTTP status code
*/
async function listBackups() {
    const path = document.getElementById('path').value;
    if (!path) {
        alert(window.appConfig.translations.fillBackupPathError);
        return;
    }

    const listBtn = document.getElementById('list-backups-btn');
    const originalText = listBtn.innerHTML;
    const tableContainer = document.getElementById('backup-files-table');

    try {
        listBtn.innerHTML = `<i class="fas fa-spinner fa-spin me-1"></i> ${window.appConfig.translations.loading}`;
        listBtn.disabled = true;

        const response = await fetch(`/list-backups`, {
            method: 'POST',
            headers: getHeaders(),
            body: JSON.stringify({ backupFilesPath: path })
        });

        const result = await response.json();

        if (!response.ok) {
            if (response.status == 401) {
                authModal.show();
            }
            throw new Error(result.errors.listBackups || result.message);
        }

        const files = result.data.backupFiles;
        if (!files || files.length === 0) {
            tableContainer.innerHTML = `<div class="alert alert-warning">${window.appConfig.translations.noBackupFilesFound}</div>`;
            return;
        }

        let tableHTML = `
            <table class="table table-hover">
                <thead>
                    <tr>
                        <th><input type="checkbox" id="select-all-backups" onchange="toggleAllBackupSelection(this)" checked></th>
                        <th>${window.appConfig.translations.backupFile}</th>
                        <th>${window.appConfig.translations.databaseName}</th>
                    </tr>
                </thead>
                <tbody>
        `;

        files.forEach(file => {
            tableHTML += `
                <tr>
                    <td><input type="checkbox" class="backup-checkbox" value="${file.fileName}" checked></td>
                    <td>${file.fileName}</td>
                    <td><input type="text" class="form-control" value="${file.defaultDbName}"></td>
                </tr>
            `;
        });

        tableHTML += `
                </tbody>
            </table>
        `;

        tableContainer.innerHTML = tableHTML;

    } catch (error) {
        console.error('Error listing backups:', error.message);
        alert(window.appConfig.translations.errorListingBackups.replace("{errorMessage}", error.message));
        tableContainer.innerHTML = `<div class="alert alert-danger">${window.appConfig.translations.errorListingBackups.replace("{errorMessage}", error.message)}</div>`;
    } finally {
        listBtn.innerHTML = originalText;
        listBtn.disabled = false;
    }
}

function toggleAllBackupSelection(source) {
    const checkboxes = document.querySelectorAll('.backup-checkbox');
    checkboxes.forEach(checkbox => {
        checkbox.checked = source.checked;
    });
}

/**
 * Returns to the last step. If the step is the latest, if changes the button class to ensure the button is pressable again
*/
function previousStep() {
    if (currentStep === totalSteps) {
        const executeBtn = document.getElementById('execute-btn');
        executeBtn.disabled = false;
        executeBtn.className = 'btn btn-success';
        executeBtn.innerHTML = `<i class="fas fa-play me-1"></i> ${window.appConfig.translations.executeOperation}`;
        document.getElementById('summary-content').innerHTML = '';
    }

    if (currentStep === 3) {
        document.querySelector('#step-3 .row').innerHTML = '';
        document.getElementById('databaseSelectButtons').innerHTML = '';
        document.getElementById('backup-path').innerHTML = '';
        document.getElementById('backup-files-table').innerHTML = '';
    }

    if (currentStep > 1) {
        currentStep--;
        updateStepNavigation();
    }
}

/**
 * Checks the current step. If something is empty it alerts to the user fill it in.
*/
function validateCurrentStep() {
    switch (currentStep) {
        case 1:
            const host = document.getElementById('host').value;
            const port = document.getElementById('port').value;
            const instance = document.getElementById('instance').value
            const user = document.getElementById('user').value;
            const password = document.getElementById('password').value;
            
            if (!host || !user || !password || (!port && !instance)) {
                alert(window.appConfig.translations.fillAllFieldsError);
                return false;
            }
            return true;
        
        case 2:
            const operation = document.querySelector('input[name="operation"]:checked');
            if (!operation) {
                alert(window.appConfig.translations.selectOperationError);
                return false;
            }
            return true;
        
        case 3:
            const ope = document.querySelector('input[name="operation"]:checked').value;
            if (ope === 'backup') {
                const databases = document.querySelectorAll('#step-3 input[type="checkbox"]:checked');
                if (databases.length === 0) {
                    alert(window.appConfig.translations.selectOneDatabaseError);
                    return false;
                }
            } else if (ope === 'restore') {
                const selectedBackups = document.querySelectorAll('#backup-files-table .backup-checkbox:checked');
                if (selectedBackups.length === 0) {
                    alert(window.appConfig.translations.selectOneBackupError);
                    return false;
                }
            }
            return true;
        
        default:
            return true;
    }
}

/**
 * Makes a GET request to the backend, collecting all the databases in the server and generating the dynamic HTML content
 * @throws {Error} Throws an error then the backend returns a bad HTTP status code
*/
async function loadDatabases() {
    const databaseContainer = document.querySelector('#step-3 .row');
    const databaseSelectButtons = document.querySelector("#databaseSelectButtons");
    
    try {
        // Shows loading
        databaseContainer.innerHTML = `
            <div class="col-12 text-center">
                <div class="spinner-border text-primary" role="status">
                    <span class="visually-hidden">${window.appConfig.translations.loading}</span>
                </div>
                <p class="mt-2">${window.appConfig.translations.loadingDatabases}</p>
            </div>
        `;

        //Fetchs the backend /databases endpoint
        const response = await fetch(`/databases`, {
            method: 'GET',
            headers: getHeaders(),
        });

        if (!response.ok) {
            const data = await response.json();
            if (response.status == 401) {
                authModal.show();
                throw new Error(`${data.message}`);
            }
            throw new Error(`${data.errors.databases}`);
        }

        const data = await response.json();
        const databases = data.data.databases;
        
        // Generate checkboxes dynamically based on backend response
        let databasesHTML = '<div class="col-md-6">';
        databases.forEach((db, index) => {
            if (index === Math.ceil(databases.length / 2)) {
                databasesHTML += '</div><div class="col-md-6">';
            }
            
            const icon = 'fas fa-database';
            const displayName = db.display_name || db.name || db;
            const value = db.name || db;
            
            databasesHTML += `
                <div class="form-check mb-3">
                    <input class="form-check-input" type="checkbox" id="db-${value}" value="${value}">
                    <label class="form-check-label" for="db-${value}">
                        <i class="${icon} me-2"></i>
                        ${displayName}
                    </label>
                </div>
            `;
        });
        databasesHTML += '</div>';

        databaseContainer.innerHTML = databasesHTML;
        databaseSelectButtons.innerHTML = `<button type="button" class="btn btn-outline-primary btn-sm" onclick="selectAllDatabases()">
                        <i class="fas fa-check-double me-1"></i>
                        ${window.appConfig.translations.selectAll}
                    </button>
                    <button type="button" class="btn btn-outline-secondary btn-sm ms-2" onclick="clearAllDatabases()">
                        <i class="fas fa-times me-1"></i>
                        ${window.appConfig.translations.clearSelection}
                    </button>`
        
    } catch (error) {
        alert(window.appConfig.translations.errorLoadingDatabases.replace("{errorMessage}", error.message));
        console.error(window.appConfig.translations.errorLoadingDatabases.replace("{errorMessage}", error.message)); 
        databaseContainer.innerHTML = `
            <div class="col-12">
                <div class="alert alert-danger">
                    <i class="fas fa-exclamation-triangle me-2"></i>
                    ${window.appConfig.translations.loadDatabasesError}
                </div>
            </div>
        `;
        
    }
}

/**
 * Checks all checkboxes in "step 3" related to databases
*/
function selectAllDatabases() {
    const checkboxes = document.querySelectorAll('#step-3 input[type="checkbox"]');
    checkboxes.forEach(checkbox => checkbox.checked = true);
}

/**
 * Clear all checkboxes in "step 3" related to databases
*/
function clearAllDatabases() {
    const checkboxes = document.querySelectorAll('#step-3 input[type="checkbox"]');
    checkboxes.forEach(checkbox => checkbox.checked = false);
}

/**
 * Generates the full sumary in the last step. Inserts the HTML content dynamically in "summary-content"
*/
function generateSummary() {
    const host = document.getElementById('host').value;
    const port = document.getElementById('port').value;
    const user = document.getElementById('user').value;
    const path = document.getElementById('path').value;
    const operation = document.querySelector('input[name="operation"]:checked').value;

    let summaryHTML = `
        <div class="summary-item">
            <div class="summary-label">
                <i class="fas fa-server me-2"></i>
                ${window.appConfig.translations.summaryDbConnection}
            </div>
            <div class="summary-value">
                <strong>${window.appConfig.translations.host}:</strong> ${host}<br>
                <strong>${window.appConfig.translations.port}:</strong> ${port}<br>
                <strong>${window.appConfig.translations.user}:</strong> ${user}<br>
                <strong>${window.appConfig.translations.password}:</strong> ••••••••
            </div>
        </div>
        
        <div class="summary-item">
            <div class="summary-label">
                <i class="fas fa-cogs me-2"></i>
                ${window.appConfig.translations.summarySelectedOperation}
            </div>
            <div class="summary-value">
                <span class="badge ${operation === 'backup' ? 'bg-success' : 'bg-warning'}">
                    ${operation === 'backup' ? window.appConfig.translations.backup : window.appConfig.translations.restore}
                </span>
            </div>
        </div>

        <div class="summary-item">
            <div class="summary-label">
                <i class="fas fa-folder-open me-2"></i>                        
                ${window.appConfig.translations.summaryBackupPath}
            </div>
            <div class="summary-value">
                ${path}<br>
            </div>
        </div>
    `;

    if (operation === 'backup') {
        const selectedDatabases = Array.from(document.querySelectorAll('#step-3 input[type="checkbox"]:checked'))
        .map(cb => cb.nextElementSibling.textContent.trim());

        summaryHTML += `<div class="summary-item">
                <div class="summary-label">
                    <i class="fas fa-database me-2"></i>
                    ${window.appConfig.translations.summarySelectedDatabases.replace("{count}", selectedDatabases.length)}
                </div>
                <div class="summary-value">
                    ${selectedDatabases.map(db => `<span class="badge bg-primary me-1 mb-1">${db}</span>`).join('')}
                </div>
            </div>`;
    } else if (operation === 'restore') {
        const selectedRows = Array.from(document.querySelectorAll('#backup-files-table .backup-checkbox:checked'));
        const selectedDbs = selectedRows.map(row => {
            const tableRow = row.closest('tr');
            return tableRow.querySelector('input[type="text"').value;
        });

        summaryHTML += `<div class="summary-item">
            <div class="summary-label">
                <i class="fas fa-database me-2"></i>
                ${window.appConfig.translations.summarySelectedDatabasesRestore.replace("{count}", selectedDbs.length)}
            </div>
            <div class="summary-value">
                ${selectedDbs.map(db => `<span class="badge bg-primary me-1 mb-1">${db}</span>`).join('')}
            </div>
        </div>`;
    }

    document.getElementById('summary-content').innerHTML = summaryHTML;
} 

// Populates the content of modalResult
function populateAndShowResultModal(operation, result) {
    const resultModal = new bootstrap.Modal(document.getElementById('resultModal'));
    const modalTitle = document.getElementById('resultModalLabel');
    const modalBody = document.getElementById('resultModalBody');

    const opCapitalized = operation.charAt(0).toUpperCase() + operation.slice(1);
    modalTitle.textContent = window.appConfig.translations.resultModalOperationSummary.replace("{operation}", opCapitalized);
    
    const completed = Array.isArray(result.data?.backupDone) || Array.isArray(result.data?.restoreDone) ? (operation === 'backup' ? result.data?.backupDone.map(db => db.name) : result.data?.restoreDone.map(db => db.database.name)) : null;
    const errors = (operation === 'backup' ? result.errors?.backupErrors : result.errors?.restoreErrors) || result.errors || null;
    const backupPath = result.data?.backupPath;      
    const totalTime = result.data?.totalTime;
    const totalOpe = (result.data?.totalBackup || result.data?.totalRestore) || null
    const totalOpeError = (result.errors?.totalBackupErrors || result.errors?.totalRestoreErrors) || null

    let statusTitle 
    if (errors && completed) {
        statusTitle = window.appConfig.translations.resultModalResultMessageCompletedWithErrors.replace("{operation}", opCapitalized);
    } else if (errors && !completed) {
        statusTitle = window.appConfig.translations.resultModalResultMessageError.replace("{operation}", opCapitalized);
    } else if (!errors && completed) {
        statusTitle = window.appConfig.translations.resultModalResultMessageSuccess.replace("{operation}", opCapitalized);
    }

    let contentHTML = `<h5 class="mb-3">${statusTitle}</h5>`;

    if (completed && completed.length > 0) {
        contentHTML += `
            <div class="result-section">
                <h6 class="result-title text-success"><i class="fas fa-check-circle me-2"></i> ${window.appConfig.translations.resultModalSuccess} ${totalOpe ? '(' + totalOpe + ')' : ''} </h6> 
                <ul class="result-list">
                    ${completed.map(db => `<li class="success">${db.name || db}</li>`).join('')}
                </ul>
            </div>
        `;
    }

    if (errors) {
        let errorItems = '';
        if (Array.isArray(errors)) {
            // Check if this is the detailed backup error structure
            if (errors.every(e => e.hasOwnProperty('Message') && e.hasOwnProperty('All') && Array.isArray(e.All) && e.All.length > 0)) {
                errorItems = errors.map(err => {
                    const mainError = err.Message;
                    const rootCause = err.All[0].Message;
                    
                    let errorMessage = `<strong>${err.DatabaseName || 'Error'}:</strong> ${mainError}`;
                    if (rootCause && rootCause !== mainError) {
                        errorMessage += ` <em>(Cause: ${rootCause})</em>`;
                    }
                    return `<li class="error">${errorMessage}</li>`;
                }).join('');
            } else if (errors.every(e => e.hasOwnProperty('database') && e.hasOwnProperty('error'))){
                errorItems = errors.map(err => {                           
                    let errorMessage = `<strong>${err.database || 'Error'}:</strong> ${err.error}`;
                    return `<li class="error">${errorMessage}</li>`;
                }).join('');
            } else {
                errorItems = errors.map(err => `<li class="error">${err}</li>`).join('');
            }
        } else if (typeof errors === 'object' && errors !== null) {
            errorItems = Object.entries(errors).map(([key, val]) => `<li class="error"><strong>${key}:</strong> ${val}</li>`).join('');
        }

        if (errorItems) {
            contentHTML += `
                <div class="result-section">
                    <h6 class="result-title text-danger"><i class="fas fa-exclamation-triangle me-2"></i> ${window.appConfig.translations.resultModalErrorsEncountered} ${totalOpeError ? '(' + totalOpeError + ')' : ''} </h6> 
                    <ul class="result-list">
                        ${errorItems}
                    </ul>
                </div>
            `;
        }
    }

    if (backupPath) {
        contentHTML += `
            <div class="result-section">
                <h6 class="result-title"><i class="fas fa-folder me-2"></i> ${window.appConfig.translations.resultModalBackupPath} </h6>
                <p><code>${backupPath}</code></p>
            </div>
        `;
    }

    if (totalTime) {
        contentHTML += `
            <div class="result-section">
                <h6 class="result-title"><i class="fas fa-folder me-2"></i> ${window.appConfig.translations.resultModalTotalTime} </h6>
                <p><code>${totalTime}</code></p>
            </div>
        `;
    }

    modalBody.innerHTML = contentHTML;
    resultModal.show();
}

/**
 * Executes the operation restore or backup.
 * If the restore operation is selected, it performs a HTTP request to /restore. 
 * If the backup operation is selected, it performs a HTTP request to /backup.
 * @throws {Error} If something did wrong in the HTTP requests.
*/
async function executeOperation() {
    const operation = document.querySelector('input[name="operation"]:checked').value;
    const selectedDatabases = Array.from(document.querySelectorAll('#step-3 input[type="checkbox"]:checked'))
        .map(cb => cb.value);
        
    const alertMsg = operation.toUpperCase() === 'BACKUP' ? 
    `${window.appConfig.translations.confirmBackup}`.replace("{operation}", operation.toUpperCase()).replace("{count}", selectedDatabases.length) :
    `${window.appConfig.translations.confirmRestore}`.replace("{operation}", operation.toUpperCase());

    if (!confirm(alertMsg)) {
        return;
    }

    const btn = document.getElementById('execute-btn');
    const prevBtn = document.getElementById('prev-btn');
    //const cancelBtn = document.getElementById('cancel-btn');
    const originalText = btn.innerHTML;
    
    btn.innerHTML = `<i class="fas fa-spinner fa-spin me-1"></i> ${window.appConfig.translations.running}`;
    btn.disabled = true;
    prevBtn.classList.add('d-none');
    //cancelBtn.classList.remove('d-none');

    try {
        let response;
        let endpoint = operation === 'backup' ? '/backup' : '/restore';
        let body;

        if (operation === 'backup') {
            const requestData = {
                databases: 
                    selectedDatabases.map(db => {
                        return {"name": db}
                    }),
                path: document.getElementById('path').value,
                concurrentOpe: parseInt(document.getElementById('maxConnections').value)
            };
            body = JSON.stringify(requestData);
        } else if (operation === 'restore') {
            const selectedRows = Array.from(document.querySelectorAll('#backup-files-table .backup-checkbox:checked'));
            const restoreData = {}
            restoreData.databases = selectedRows.map(row => {
                const tableRow = row.closest('tr');
                const backupFileName = row.value;
                const dbName = tableRow.querySelector('input[type="text"').value;
                const backupPath = document.getElementById('path').value;
                const fullPath = backupPath.endsWith('/') || backupPath.endsWith('\\') ? backupPath : backupPath + '/';

                return {
                    name: dbName,
                    backupPath: fullPath + backupFileName,
                };
            })
            restoreData.concurrentOpe = parseInt(document.getElementById('maxConnections').value);
            console.log("restoreData: ", JSON.stringify(restoreData));
            body = JSON.stringify(restoreData);
            console.log(body);
        }

        response = await fetch(endpoint, {
            method: 'POST',
            headers: getHeaders(),
            body: body
        });          
        
        const result = await response.json();
        
        if (!response.ok) {
            throw result;
        }

        populateAndShowResultModal(operation, result);
        
    } catch (error) {
        populateAndShowResultModal(operation, error);
    } finally {
        btn.innerHTML = originalText;
        btn.disabled = false;
        prevBtn.classList.remove('d-none');
        //cancelBtn.classList.add('d-none');
    }
}

/**
 * Emptys the session
*/
async function logout() {
    try {
        response = await fetch(`/logout`, {
                            method: "GET",
                            headers: getHeaders()
                        }
                    );

        if (!response.ok) {
            throw new Error(`Error trying to fetch the URL: ${window.location.href}`);
        }
        
        updateAuthStatus(false);
        authModal.show();

    } 
    catch(err) {
        console.error(`Error trying to logout: ${err}`)
        alert(`Error trying to logout: ${err}`);
    }
    
    
}

/**
 * Check if the user is authenticated. At first, checks if the 'jwt' cookie is set, then checks if the jwt is valid
 * @returns {boolean}
*/
async function isAuth(){
    try {
        response = await fetch(`/session`, {
            method: "GET",
            headers: getHeaders()
        });
        if (!response.ok) {
            return false;
        }
        data = await response.json();
        return data;
    }
    catch (err){
        console.error(`Error while trying to get session: ${err}`);
        alert(`Error while trying to get session: ${err}`);
    }
    
}

/**
 * Creates a headers object for fetch requests, including Content-Type,
 * the CSRF token read from the meta tag, and the Authorization token if needed.
 * @returns {Object} The headers object.
*/
function getHeaders() {
    const csrfToken = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
    
    const headers = {
        'Content-Type': 'application/json',
    };

    if(window.appConfig.appCSRFTokenUsage) {
        headers['X-CSRF-TOKEN'] = csrfToken;
    }
    return headers;
}

function togglePortInstance(event) {
    event.preventDefault();
    const portGroup = document.getElementById('port-field-group');
    const instanceGroup = document.getElementById('instance-field-group');
    const portInput = document.getElementById('port');
    const instanceInput = document.getElementById('instance');

    if (portGroup.classList.contains('d-none')) {
        portGroup.classList.remove('d-none');
        instanceGroup.classList.add('d-none');
        instanceInput.value = '';
        portInput.setAttribute('required', '');
        instanceInput.removeAttribute('required');
    } else {
        portGroup.classList.add('d-none');
        instanceGroup.classList.remove('d-none');
        portInput.value = '';
        instanceInput.setAttribute('required', '');
        portInput.removeAttribute('required');
    }
}

//Starts GUI
updateStepNavigation();
