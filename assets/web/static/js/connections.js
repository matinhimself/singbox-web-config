// Connections Manager
class ConnectionsManager {
    constructor() {
        this.ws = null;
        this.connections = [];
        this.filteredConnections = [];
        this.filters = {
            search: '',
            network: '',
            source: '',
            chain: ''
        };
        this.sortBy = 'start-desc';
        this.selectedConnection = null;
        this.outbounds = [];
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;

        this.init();
    }

    init() {
        this.setupEventListeners();
        this.connectWebSocket();
        this.loadOutbounds();
    }

    setupEventListeners() {
        // Search input
        document.getElementById('search-input').addEventListener('input', (e) => {
            this.filters.search = e.target.value.toLowerCase();
            this.applyFiltersAndSort();
        });

        // Filter selects
        document.getElementById('filter-network').addEventListener('change', (e) => {
            this.filters.network = e.target.value;
            this.applyFiltersAndSort();
        });

        document.getElementById('filter-source').addEventListener('change', (e) => {
            this.filters.source = e.target.value;
            this.applyFiltersAndSort();
        });

        document.getElementById('filter-chain').addEventListener('change', (e) => {
            this.filters.chain = e.target.value;
            this.applyFiltersAndSort();
        });

        // Sort select
        document.getElementById('sort-by').addEventListener('change', (e) => {
            this.sortBy = e.target.value;
            this.applyFiltersAndSort();
        });

        // Clear filters button
        document.getElementById('clear-filters').addEventListener('click', () => {
            this.clearFilters();
        });

        // Rule checkboxes - update preview
        const checkboxes = document.querySelectorAll('.rule-checkbox');
        checkboxes.forEach(cb => {
            cb.addEventListener('change', () => this.updateRulePreview());
        });

        // Outbound select - update preview
        document.getElementById('rule-outbound').addEventListener('change', () => {
            this.updateRulePreview();
        });
    }

    connectWebSocket() {
        const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${wsProtocol}//${window.location.host}/ws/connections`;

        this.updateStatus('Connecting...', 'connecting');

        try {
            this.ws = new WebSocket(wsUrl);

            this.ws.onopen = () => {
                console.log('WebSocket connected');
                this.updateStatus('Connected', 'connected');
                this.reconnectAttempts = 0;
            };

            this.ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleConnectionsUpdate(data);
                } catch (error) {
                    console.error('Failed to parse WebSocket message:', error);
                }
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.updateStatus('Connection Error', 'error');
            };

            this.ws.onclose = () => {
                console.log('WebSocket closed');
                this.updateStatus('Disconnected', 'disconnected');
                this.attemptReconnect();
            };
        } catch (error) {
            console.error('Failed to create WebSocket:', error);
            this.updateStatus('Connection Failed', 'error');
        }
    }

    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
            console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
            setTimeout(() => this.connectWebSocket(), delay);
        } else {
            this.updateStatus('Connection Lost', 'error');
        }
    }

    updateStatus(text, status) {
        const statusEl = document.getElementById('ws-status');
        statusEl.textContent = text;
        statusEl.className = `connection-status status-${status}`;
    }

    handleConnectionsUpdate(data) {
        this.connections = data.connections || [];

        // Update stats
        this.updateStats(data);

        // Update filter options
        this.updateFilterOptions();

        // Apply filters and render
        this.applyFiltersAndSort();
    }

    updateStats(data) {
        document.getElementById('stat-count').textContent = (data.connections || []).length;
        document.getElementById('stat-upload').textContent = this.formatBytes(data.uploadTotal || 0);
        document.getElementById('stat-download').textContent = this.formatBytes(data.downloadTotal || 0);
        document.getElementById('stat-memory').textContent = this.formatBytes(data.memory || 0);
    }

    updateFilterOptions() {
        // Get unique sources and chains
        const sources = new Set();
        const chains = new Set();

        this.connections.forEach(conn => {
            if (conn.metadata && conn.metadata.sourceIP) {
                sources.add(conn.metadata.sourceIP);
            }
            if (conn.chains && conn.chains.length > 0) {
                conn.chains.forEach(chain => chains.add(chain));
            }
        });

        // Update source filter
        const sourceSelect = document.getElementById('filter-source');
        const currentSource = sourceSelect.value;
        sourceSelect.innerHTML = '<option value="">All Sources</option>';
        Array.from(sources).sort().forEach(source => {
            const option = document.createElement('option');
            option.value = source;
            option.textContent = source;
            if (source === currentSource) option.selected = true;
            sourceSelect.appendChild(option);
        });

        // Update chain filter
        const chainSelect = document.getElementById('filter-chain');
        const currentChain = chainSelect.value;
        chainSelect.innerHTML = '<option value="">All Chains</option>';
        Array.from(chains).sort().forEach(chain => {
            const option = document.createElement('option');
            option.value = chain;
            option.textContent = chain;
            if (chain === currentChain) option.selected = true;
            chainSelect.appendChild(option);
        });
    }

    applyFiltersAndSort() {
        // Apply filters
        this.filteredConnections = this.connections.filter(conn => {
            // Search filter
            if (this.filters.search) {
                const searchText = JSON.stringify(conn).toLowerCase();
                if (!searchText.includes(this.filters.search)) {
                    return false;
                }
            }

            // Network filter
            if (this.filters.network && conn.metadata.network !== this.filters.network) {
                return false;
            }

            // Source filter
            if (this.filters.source && conn.metadata.sourceIP !== this.filters.source) {
                return false;
            }

            // Chain filter
            if (this.filters.chain && !conn.chains.includes(this.filters.chain)) {
                return false;
            }

            return true;
        });

        // Apply sorting
        this.sortConnections();

        // Render
        this.renderConnections();
    }

    sortConnections() {
        const [field, direction] = this.sortBy.split('-');

        this.filteredConnections.sort((a, b) => {
            let aVal, bVal;

            switch (field) {
                case 'start':
                    aVal = new Date(a.start);
                    bVal = new Date(b.start);
                    break;
                case 'download':
                    aVal = a.download || 0;
                    bVal = b.download || 0;
                    break;
                case 'upload':
                    aVal = a.upload || 0;
                    bVal = b.upload || 0;
                    break;
                case 'source':
                    aVal = a.metadata.sourceIP || '';
                    bVal = b.metadata.sourceIP || '';
                    break;
                case 'destination':
                    aVal = a.metadata.destinationIP || '';
                    bVal = b.metadata.destinationIP || '';
                    break;
                default:
                    return 0;
            }

            if (direction === 'asc') {
                return aVal > bVal ? 1 : aVal < bVal ? -1 : 0;
            } else {
                return aVal < bVal ? 1 : aVal > bVal ? -1 : 0;
            }
        });
    }

    renderConnections() {
        const tbody = document.getElementById('connections-tbody');

        if (this.filteredConnections.length === 0) {
            tbody.innerHTML = `
                <tr class="empty-state-row">
                    <td colspan="10">
                        <div class="empty-state">
                            <p>No connections found</p>
                        </div>
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = this.filteredConnections.map(conn => this.renderConnectionRow(conn)).join('');
    }

    renderConnectionRow(conn) {
        const duration = this.formatDuration(new Date(conn.start));
        const chains = (conn.chains || []).join(' → ');
        const host = conn.metadata.host || '-';

        return `
            <tr class="connection-row" data-id="${conn.id}">
                <td>
                    <div class="connection-address">
                        ${conn.metadata.sourceIP}:${conn.metadata.sourcePort}
                    </div>
                </td>
                <td>
                    <div class="connection-address">
                        ${conn.metadata.destinationIP}:${conn.metadata.destinationPort}
                    </div>
                </td>
                <td>
                    <span class="network-badge network-${conn.metadata.network}">
                        ${conn.metadata.network.toUpperCase()}
                    </span>
                </td>
                <td class="connection-host">${this.truncate(host, 30)}</td>
                <td class="connection-chains">${this.truncate(chains, 40)}</td>
                <td class="traffic-cell">${this.formatBytes(conn.upload)}</td>
                <td class="traffic-cell">${this.formatBytes(conn.download)}</td>
                <td>${duration}</td>
                <td class="rule-cell">${this.truncate(conn.rule || '-', 50)}</td>
                <td>
                    <button class="button button-small button-primary"
                            onclick="connectionsManager.openRuleModal('${conn.id}')">
                        + Rule
                    </button>
                </td>
            </tr>
        `;
    }

    openRuleModal(connectionId) {
        const conn = this.connections.find(c => c.id === connectionId);
        if (!conn) return;

        this.selectedConnection = conn;

        // Populate modal with connection data
        document.getElementById('rule-source-ip-value').textContent = conn.metadata.sourceIP;
        document.getElementById('rule-destination-ip-value').textContent = conn.metadata.destinationIP;
        document.getElementById('rule-destination-port-value').textContent = conn.metadata.destinationPort;
        document.getElementById('rule-network-value').textContent = conn.metadata.network;
        document.getElementById('rule-domain-value').textContent = conn.metadata.host || 'N/A';

        // Disable domain checkbox if no host
        const domainCheckbox = document.getElementById('rule-domain');
        domainCheckbox.disabled = !conn.metadata.host;
        if (!conn.metadata.host) {
            domainCheckbox.checked = false;
        }

        // Reset checkboxes
        document.querySelectorAll('.rule-checkbox').forEach(cb => {
            if (!cb.disabled) cb.checked = false;
        });

        // Reset outbound
        document.getElementById('rule-outbound').value = '';

        // Update preview
        this.updateRulePreview();

        // Show modal
        document.getElementById('rule-modal').style.display = 'flex';
    }

    updateRulePreview() {
        if (!this.selectedConnection) return;

        const rule = {};
        const conn = this.selectedConnection;

        if (document.getElementById('rule-source-ip').checked) {
            rule.source_ip_cidr = [conn.metadata.sourceIP + '/32'];
        }

        if (document.getElementById('rule-destination-ip').checked) {
            rule.ip_cidr = [conn.metadata.destinationIP + '/32'];
        }

        if (document.getElementById('rule-destination-port').checked) {
            rule.port = [conn.metadata.destinationPort];
        }

        if (document.getElementById('rule-network').checked) {
            rule.network = [conn.metadata.network];
        }

        if (document.getElementById('rule-domain').checked && conn.metadata.host) {
            rule.domain_suffix = [conn.metadata.host];
        }

        const outbound = document.getElementById('rule-outbound').value;
        if (outbound) {
            rule.outbound = outbound;
        }

        document.getElementById('rule-preview-json').textContent = JSON.stringify(rule, null, 2);
    }

    async loadOutbounds() {
        try {
            const response = await fetch('/api/config/export');
            const config = await response.json();

            this.outbounds = [];
            if (config.outbounds) {
                config.outbounds.forEach(outbound => {
                    if (outbound.tag) {
                        this.outbounds.push(outbound.tag);
                    }
                });
            }

            // Populate outbound select
            const select = document.getElementById('rule-outbound');
            select.innerHTML = '<option value="">Select outbound...</option>';
            this.outbounds.forEach(tag => {
                const option = document.createElement('option');
                option.value = tag;
                option.textContent = tag;
                select.appendChild(option);
            });
        } catch (error) {
            console.error('Failed to load outbounds:', error);
        }
    }

    clearFilters() {
        this.filters = {
            search: '',
            network: '',
            source: '',
            chain: ''
        };

        document.getElementById('search-input').value = '';
        document.getElementById('filter-network').value = '';
        document.getElementById('filter-source').value = '';
        document.getElementById('filter-chain').value = '';

        this.applyFiltersAndSort();
    }

    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
    }

    formatDuration(startTime) {
        const now = new Date();
        const diff = now - startTime;
        const seconds = Math.floor(diff / 1000);

        if (seconds < 60) return `${seconds}s`;
        const minutes = Math.floor(seconds / 60);
        if (minutes < 60) return `${minutes}m ${seconds % 60}s`;
        const hours = Math.floor(minutes / 60);
        return `${hours}h ${minutes % 60}m`;
    }

    truncate(str, maxLen) {
        if (!str) return '-';
        if (str.length <= maxLen) return str;
        return str.substring(0, maxLen - 3) + '...';
    }
}

// Initialize connections manager
let connectionsManager;
document.addEventListener('DOMContentLoaded', () => {
    connectionsManager = new ConnectionsManager();
});

// Modal functions (global scope for onclick handlers)
function closeRuleModal() {
    document.getElementById('rule-modal').style.display = 'none';
    connectionsManager.selectedConnection = null;
}

async function createRuleFromConnection() {
    if (!connectionsManager.selectedConnection) return;

    const conn = connectionsManager.selectedConnection;
    const formData = new FormData();

    if (document.getElementById('rule-source-ip').checked) {
        formData.append('source_ip', conn.metadata.sourceIP);
    }

    if (document.getElementById('rule-destination-ip').checked) {
        formData.append('destination_ip', conn.metadata.destinationIP);
    }

    if (document.getElementById('rule-destination-port').checked) {
        formData.append('destination_port', conn.metadata.destinationPort);
    }

    if (document.getElementById('rule-network').checked) {
        formData.append('network', conn.metadata.network);
    }

    if (document.getElementById('rule-domain').checked && conn.metadata.host) {
        formData.append('domain', conn.metadata.host);
    }

    const outbound = document.getElementById('rule-outbound').value;
    if (!outbound) {
        alert('Please select an outbound');
        return;
    }
    formData.append('outbound', outbound);

    try {
        const response = await fetch('/api/connections/create-rule', {
            method: 'POST',
            body: formData
        });

        if (response.ok) {
            alert('Rule created successfully!');
            closeRuleModal();
        } else {
            const error = await response.text();
            alert('Failed to create rule: ' + error);
        }
    } catch (error) {
        console.error('Error creating rule:', error);
        alert('Failed to create rule: ' + error.message);
    }
}
