# Spécifications Complètes - Système de Surveillance par Drones
## Version 1.0 - Novembre 2024

# Table des Matières
1. [Objectifs et Vue d'Ensemble](#1-objectifs-et-vue-densemble)
2. [Architecture du Système](#2-architecture-du-système)
3. [Modèle de Simulation](#3-modèle-de-simulation)
4. [Implémentation Technique](#4-implémentation-technique)
5. [Métriques et Évaluation](#5-métriques-et-évaluation)
6. [Plan de Test](#6-plan-de-test)
7. [Configuration et Paramètres](#7-configuration-et-paramètres)
8. [Bibliographie](#8-bibliographie)

## 1. Objectifs et Vue d'Ensemble

### 1.1 Objectif Principal
Développer une simulation en Go pour évaluer l'efficacité des protocoles de surveillance par drones lors d'événements publics, utilisant une approche distribuée et des communications P2P.

### 1.2 Sous-Objectifs
- Optimiser la détection des incidents
- Minimiser le temps de réponse
- Maximiser la couverture spatiale
- Évaluer différents protocoles de surveillance

### 1.3 Contexte
Surveillance d'événements publics avec contraintes :
- Durée : 2-8 heures
- Participants : 1,000-10,000
- Surface : Variable selon configuration

[Source: [8], [10]]

## 2. Architecture du Système

### 2.1 Environnement de Simulation

#### 2.1.1 Structure Spatiale
- **Grille 2D**
  - Taille cellule : 5m x 5m
  - Coordonnées réelles pour positionnement précis
  - Capacité maximale : 6 personnes/m²

#### 2.1.2 Zones Définies
- **Entrées/Sorties**
  - 4 points aux extrémités
  - Capacité : 50 personnes/minute
- **Zone Centrale (Scène)**
  - Surface définie : ~2000m²
  - Densité maximale plus élevée
- **Zones Périphériques**
  - Restauration
  - Repos
  - Services
- **Stations de Recharge Drones**
  - Points fixes
  - Capacité : 2 drones simultanés

[Source: [1], [9], [10]]

### 2.2 Composants du Système

#### 2.2.1 Caractéristiques des Drones

**Spécifications Techniques**
```yaml
Performance:
  Autonomie: 45 minutes (charge)
  Temps_recharge: 40 minutes (80%)
  Hauteur_optimale: 40m
  Vitesse_max: 15 m/s
  Accélération_max: 3 m/s²
```

**Capacités de Détection**
```yaml
Détection:
  Champ_vision: 84° horizontal
  Précision_centre: 95%
  Décroissance: exponentielle
  Fréquence_scan: 10 Hz
```

[Source: [2], [9]]

#### 2.2.2 Modèle des Individus

**États Possibles**
- Normal (debout)
- Malaise (allongé)
- Critique (urgence)

**Modèle de Probabilité de Malaise**
```python
P(malaise) = P_base * (1 + Σ facteurs_risque)
où:
P_base = 0.001 par personne/heure
```

**Facteurs de Risque**
```yaml
Facteurs:
  Température_30C+: +0.5
  Densité_4p/m²: +0.3
  Durée_6h+: +0.2
  Proximité_scène: +0.4
```

[Source: [5], [6], [7]]

## 3. Modèle de Simulation

### 3.1 Capacités de Détection et Couverture

#### 3.1.1 Modèle de Couverture Spatiale
```python
# Calcul de la surface couverte
Rayon_couverture = hauteur * tan(angle_vision/2)
Surface_couverte = π * Rayon_couverture²

# Pour h=40m, angle=84°
Rayon_effectif = 37.2m
Surface_couverte = 4,347m²
Cellules_couvertes = ~174 cellules
```

[Source: [1], [2]]

#### 3.1.2 Modèle de Précision
```python
Précision(d) = Précision_max * exp(-α * (d/R_max)²)
où:
- Précision_max = 0.95
- α = 2.3
- d = distance au centre
- R_max = rayon maximum
```

[Source: [2]]

### 3.2 Protocoles de Surveillance

#### 3.2.1 Protocole Standard (PS)
1. **Couverture Systématique**
   - Découpage en zones régulières
   - Rotation selon pattern prédéfini
   - Temps fixe par zone

2. **Détection**
   - Scan continu à 10 Hz
   - Confirmation multi-angle
   - Seuil de confiance : 85%

#### 3.2.2 Protocole Adaptatif (PA)
1. **Évaluation des Risques**
   ```python
   Risque(cellule) = (
       Base +
       Densité * 0.3 +
       Historique * 0.2 +
       Proximité_scène * 0.3
   )
   ```

2. **Allocation Dynamique**
   - Actualisation : 60 secondes
   - Prioritisation zones à risque
   - Maintien couverture minimale

#### 3.2.3 Protocole Collaboratif (PC)
1. **Communication**
   - Broadcast état : 5 secondes
   - Partage immédiat détections
   - Consensus distribué

2. **Optimisation**
   ```python
   Score_global = Σ(Couverture_i * Qualité_détection_i)
   ```

[Source: [3], [4]]

## 4. Implémentation Technique

### 4.1 Structures de Données

```go
// Structure principale
type Simulation struct {
    Grid        *Grid
    Drones      []*Drone
    Agents      []*Agent
    Metrics     *Metrics
    Config      *Config
}

// Représentation spatiale
type Grid struct {
    Cells       [][]*Cell
    Updates     chan UpdateMsg
    mutex       sync.RWMutex
}

type Cell struct {
    Occupants    []Agent
    Density      float64
    RiskScore    float64
    mutex        sync.RWMutex
}

// Composants actifs
type Drone struct {
    Position    Position3D
    Coverage    [][]float64
    Battery     float64
    Channel     chan DroneMsg
    State       DroneState
}

type Agent struct {
    Position    Position2D
    State       AgentState
    TimeOnSite  time.Duration
}
```

### 4.2 Système de Communication

```go
// Messages
type DroneMsg struct {
    Type        MsgType
    Position    Position3D
    Detection   *Detection
    Timestamp   time.Time
}

// Communication P2P
func (d *Drone) Broadcast(msg DroneMsg) {
    for _, neighbor := range d.Neighbors {
        select {
        case neighbor.Channel <- msg:
        default:
            metrics.IncrementDroppedMessages()
        }
    }
}
```

### 4.3 Cycle de Mise à Jour

```go
func (s *Simulation) Update() {
    // 1. Update Agents
    s.updateAgents()
    
    // 2. Update Drones
    s.updateDrones()
    
    // 3. Process Detections
    s.processDetections()
    
    // 4. Update Metrics
    s.updateMetrics()
}

// Exemple d'update des agents
func (s *Simulation) updateAgents() {
    for _, agent := range s.Agents {
        s.Grid.mutex.Lock()
        oldCell := s.Grid.GetCell(agent.Position)
        newPos := agent.UpdatePosition()
        newCell := s.Grid.GetCell(newPos)
        
        oldCell.RemoveAgent(agent)
        newCell.AddAgent(agent)
        s.Grid.mutex.Unlock()
    }
}
```

[Source: [4], [8]]

## 5. Métriques et Évaluation

### 5.1 Métriques en Temps Réel

```go
type Metrics struct {
    // Couverture
    CoverageRate       float64    // % zone couverte
    CoverageQuality    float64    // qualité moyenne
    
    // Détection
    DetectionRate      float64    // % incidents détectés
    FalsePositives     int
    FalseNegatives     int
    
    // Performance
    ResponseTime       []float64  // temps de réponse
    BatteryEfficiency  float64
    DroneUtilization   float64
}
```

### 5.2 Calcul de Performance

```python
Performance_globale = (
    0.4 * Taux_détection +
    0.3 * Temps_réponse_normalisé +
    0.2 * Couverture_moyenne +
    0.1 * Efficacité_batterie
)
```

[Source: [8]]

## 6. Plan de Test

### 6.1 Tests Unitaires

```go
func TestDetection(t *testing.T) {
    tests := []struct {
        name     string
        distance float64
        expected float64
    }{
        {"CentreZone", 0.0, 0.95},
        {"MiDistance", 18.6, 0.75},
        {"LimiteCouverture", 37.2, 0.50},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculatePrecision(tt.distance)
            if math.Abs(result-tt.expected) > 0.01 {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### 6.2 Scénarios de Test

#### 6.2.1 Scénario de Base
- 1000 agents
- 4 drones
- 2 heures simulation
- Conditions normales

#### 6.2.2 Scénario Haute Densité
- 5000 agents
- 8 drones
- Zones congestionnées

#### 6.2.3 Scénario Dégradé
- Perte de drones
- Communications perturbées
- Conditions météo défavorables

### 6.3 Validation

Pour chaque scénario:
1. Exécuter 30 itérations
2. Calculer:
   - Moyenne
   - Écart-type
   - Intervalles de confiance
3. Analyser:
   - Stabilité
   - Performance
   - Ressources

[Source: [8]]

## 7. Configuration et Paramètres

```yaml
simulation:
  tick_rate: 100ms
  grid_size: [100, 100]
  cell_size: 5m
  duration: 2h

drones:
  count: 4
  height: 40m
  battery_capacity: 45min
  detection_rate: 10Hz
  max_speed: 15
  acceleration: 3

detection:
  base_precision: 0.95
  decay_factor: 2.3
  min_confidence: 0.85
  scan_frequency: 10

communication:
  broadcast_rate: 5s
  message_timeout: 100ms
  buffer_size: 100

crowd:
  max_density: 6
  base_malaise_prob: 0.001
  risk_factors:
    temperature: 0.5
    density: 0.3
    duration: 0.2
    proximity: 0.4
```

## 8. Bibliographie

[1] "UAV Coverage Optimization for Urban Surveillance", Robotics and Automation Letters, 2023

[2] "Spatial Accuracy Models in UAV Surveillance", Sensors Journal IEEE, 2023

[3] "Adaptive UAV Patrol Strategies", Autonomous Robots, 2023

[4] "Distributed UAV Coordination Protocols", ICRA 2023

[5] "Medical Incidents at Outdoor Music Festivals", Prehospital and Disaster Medicine, 2022

[6] "Risk Factors for Medical Emergencies at Large Public Events", International Journal of Environmental Research and Public Health, 2021

[7] "Analysis of Medical Interventions at Music Festivals", Scandinavian Journal of Trauma, 2023

[8] "Validation Protocols for Multi-Agent Simulations", Simulation Modelling Practice and Theory, 2023

[9] "Professional Drone Operations and Maintenance", IEEE Aerospace Conference, 2022

[10] "Crowd Dynamics and Safety at Mass Events", Safety Science Journal, 2023
