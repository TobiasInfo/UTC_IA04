# Spécifications Complètes - Système de Surveillance par Drones
## Version 1.0 - Novembre 2024

# Table des Matières
1. [Objectifs et Vue d'Ensemble](#1-objectifs-et-vue-densemble)  
2. [Lancement de la simulation](#2-lancement-de-la-simulation)  
3. [Architecture du Système](#3-architecture-du-système)  
4. [Implémentation](#4-implémentation)  
5. [Métriques et Évaluation](#5-métriques-et-évaluation)  
6. [Bibliographie](#6-bibliographie)  

## 1. Objectifs et Vue d'Ensemble

### 1.1 Objectif Principal
Développer une simulation en Go pour évaluer l'efficacité des protocoles de surveillance par drones lors d'événements publics, utilisant une approche distribuée et des communications P2P.

### 1.2 Sous-Objectifs
- Pouvoir modifier l'agencement des éléments de la carte  
- Pouvoir modifier le nombre de personnes et de drones présents  
- Obtenir des informations sur la simulation pour évaluer la performance des différents protocoles de surveillance  

### 1.3 Contexte
Surveillance d'événements publics avec contraintes :  
- Durée : 2-8 heures  
- Participants : 1,000-10,000  
- Agencement : Variable selon configuration  

[Source: [8], [10]]

## 2. Lancement de la simulation

### 2.1 Ouvrir l'application

Pour lancer la simulation, veuillez :  
1. Télécharger le répertoire accessible sur https://github.com/TobiasInfo/UTC_IA04  
2. Accéder au répertoire en ligne de commande : cd XXX\UTC_IA04\cmd  
3. Exécuter en ligne de commande le fichier main : go run .\main_gui_ebiten.go  

### 2.2 Paramétrage

Dans le Menu principal peuvent être choisis :  
1. Le nombre de drones d'observation  
2. Le nombre de festivaliers  
3. La carte parmi celles proposées  
4. Le protocole de communication et d'observation des drones  

Une fois ces paramètres choisis, cliquer sur "Start Simulation".  
Le mode "Debug" désactive l'évolution automatique du système, et impose à l'utilisateur d'utiliser le bouton "Update Simulation" à chaque étape.

### 2.3 Fonctionnalités

Une fois la simulation lancée, elle peut à tout moment être mise en pause à l'aide du bouton "Pause". Il est également possible une fois la simulation pausée de la mettre à jour manuellement avec "Update Simulation" pour que le système évolue au rythme voulu.

Des informations sur la simulation sont affichées sur le bandeau inférieur et rendent compte de l'état actuel de la simulation.  
Des graphiques représentant l'évolution au cours du temps des informations clés sont sauvegardés dans le fichier UTC_IA04 à la fin de la simulation. La simulation prend fin lorsque tous les participants ont quitté le festival, et que les drones sont allés se poser.

Au cours de la simulation des graphes sont également affichés, celui de gauche représente la densité de participants sur la carte, pour savoir où les festivaliers sont le plus présents. Celui de droite représente les drones, leur champ de communication ainsi que les réseaux de communication formés. Il est également possible de lire qu'un drone est entré en communication avec la tente de secours. Ces graphiques peuvent être agrandis ou réduits selon la convenance en cliquant dessus.

Il est possible d'obtenir des informations supplémentaires sur un participant, un drone ou un point d'intérêt en le survolant avec la souris. Attention, lorsque trop d'agents se superposent, l'info-bulle perd en lisibilité.

### 2.4 Lecture de l'interface

La fenêtre représente un festival, où l'on trouve des points d'intérêts pour les festivaliers (que l'on appellera par la suite POI), des participants, des drones, et des POI pour les drones.  
L'entrée du festival se trouve sur la gauche de la fenêtre tandis que la sortie se trouve à droite. Les festivaliers ne peuvent pas sortir de la carte centrale, sauf dans le cas d'une sortie définitive du festival.  
Les drones survolent les participants, et l'on peut observer la portée de vision d'un drone avec le disque d'ombre autour de lui.  
Les POI des participants sont représentés selon leur fonction (scène, aire de repos, stand de nourriture, toilettes,...).  
Les POI des drones (station de recharge et tente infirmerie) sont représentés avec des images différentes.  
Les sauveteurs sont représentés avec une image de pompier, ils sont reliés par une ligne verte à leur poste de secours et à la position de la personne qu'ils vont sauver.

## 3. Architecture du Système

### 3.1 Environnement de Simulation

- **Plan 2D+**  
  - Coordonnées réelles pour les participants  
  - Grille de 30 sur 20 pour les drones et les POI  
  - Les participants ne peuvent pas traverser les POI, les drones les survolent  

- **Entrée/Sortie**  
  - Entrée sur la gauche de l'écran et sortie sur la droite  
  - Phases : Les premières minutes permettent à nouveaux participants d'arriver, mais les entrées sont ensuite fermées. Tout au long de la simulation les participants peuvent sortir, mais lorsque le temps est écoulé, les participants sont obligés de se diriger vers la sortie.

[Source: [1], [9], [10]]

### 3.2 Composants du Système

#### 3.2.1 Caractéristiques des Drones

**Paramètres d'un drone**  
- Batterie  
- Portée de Communication  
- Champ de Vision  
- Protocole  

**Champ de Vision**  
Pour chaque participant en situation de détresse dans le champ de vision d'un drone, la probabilité qu'il soit identifié correctement dépend de sa distance par rapport au drone et de la quantité de personnes dans le champ de vision du drone et est calculée :
```go
probaDetection := max(0, 1.0/float64(s.DroneSeeRange)-(float64(nbPersDetected)*0.03))
```

**Protocole 1**

Step 1 :  
- Je scanne les personnes en danger  
- Si je vois une personne en danger, je la sauvegarde.

Step 2 :  
- Dès que ma liste est supérieure > 1 je m'en vais vers le RP le + Proche pour régler les problèmes.  
- Si je n'ai plus de batterie, je bouge vers le point de charge le plus proche.  
    - J'essaye lors de mon mouvement de transmettre ma liste à mes voisins pour qu'ils aillent informer le rescuer à ma place.  
- Une fois que ma charge est terminée, je bouge vers le point de sauvetage le plus proche.

**Protocole 2**

Step 0 :  
- Si je n'ai plus de batterie, je bouge vers le point de charge le plus proche.  
    - J'essaye lors de mon mouvement de transmettre ma liste à mes voisins pour qu'ils aillent informer le rescuer à ma place.  
- Une fois que ma charge est terminée, je bouge vers le point de sauvetage le plus proche.

Step 1 :  
- Je scanne les personnes en danger  
- Si je vois une personne en danger, je la sauvegarde.

Step 2 :  
- J'essaye de communiquer avec un RP si un RP est dans mon rayon de communication.  
   - Si aucun RP n'est dans mon rayon de communication.  
		- J'essaye de voir si je peux envoyer l'information à un drone qui est en n+1 de mon rayon de communication.  
		- Si je ne peux pas, je bouge vers le rescue point le plus proche.  
- Je bouge vers le rescue point si je ne peux pas communiquer.

**Protocole 3**

Step 0 :  
- Si je n'ai plus de batterie, je bouge vers le point de charge le plus proche.  
    - J'essaye lors de mon mouvement de transmettre ma liste à mes voisins pour qu'ils aillent informer le rescuer à ma place.  
- Une fois que ma charge est terminée, je bouge vers le point de sauvetage le plus proche.

Step 1 :  
- Je scanne les personnes en danger  
- Si je vois une personne en danger, je la sauvegarde.

Step 2 :  
- J'essaye de communiquer avec un RP si un RP est dans mon rayon de communication.  
   - Si aucun RP n'est dans mon rayon de communication.  
		- J'essaye de voir si je peux envoyer l'information à un drone qui est dans mon network.  
			- Un network est un sous-ensemble de drones qui peuvent communiquer entre eux, ils sont chaînés et ils forment un sous-graphe.  
		- Si je ne peux pas, je bouge vers le rescue point le plus proche.  
- Je bouge vers le rescue point si je ne peux pas communiquer.

**Protocole 4**

Step 0 :  
- Si je n'ai plus de batterie, je bouge vers le point de charge le plus proche.  
    - J'essaye lors de mon mouvement de transmettre ma liste à mes voisins pour qu'ils aillent informer le rescuer à ma place.  
- Une fois que ma charge est terminée, je bouge vers le point de sauvetage le plus proche.

Step 1 :  
- Je scanne les personnes en danger  
- Si je vois une personne en danger, je la sauvegarde.

Step 2 :  
- J'essaye de communiquer avec un RP si un RP est dans mon rayon de communication.  
   - Si aucun RP n'est dans mon rayon de communication.  
		- J'essaye de voir si je peux envoyer l'information à un drone qui est dans mon network.  
			- Un network est un sous-ensemble de drones qui peuvent communiquer entre eux, ils sont chaînés et ils forment un sous-graphe.  
		- Si je ne peux pas, je prends le drone le plus proche dans mon network en termes de distance d'un RP et je lui transfère la responsabilité de sauver les personnes.

Step 3 :  
- Je bouge vers le rescue point si je suis le drone le plus proche.

[Source: = [3], [4], [2], [9]]

#### 3.2.2 Modèle des Participants

**États Possibles**  
- Normal (debout) - Consommation faible d'énergie  
- Repos - Récupération d'énergie  
- Malaise (allongé) - Consommation rapide d'énergie  

**Modèle de Probabilité de Malaise**
```python
P(malaise) = P_base x (1 - Resistance au Malaise du participant) * (1 - Energie du participant)
où:
P_base = 0.005
```

**Intérêts**  
4 profils de participants :  
- Adventurous  
- Cautious  
- Social  
- Independent  

Ces profils ont une influence sur :  
- La vitesse de déplacement de l'individu  
- les variables "CrowdFollowingTendency" et "PersonalSpace" qui influe sur la tendance de l'individu à aller ou non dans les zones avec de nombreux autres individus  
- Le niveau d'énergie à partir duquel un individu passera en mode repos  
- La résistance au malaise de l'individu  
- L'intérêt porté par l'individu à chaque POI, et donc vers lesquels il préférera se diriger.

Lorsque qu'un participant atteint un POI, il va y rester pendant une durée variable, puis repartir à la recherche d'un autre POI.

#### 3.2.3 Sauvetage d'un participant en détresse

Lorsqu'un participant en situation de détresse a été remarqué, l'information doit être remontée à la tente infirmerie qui détachera ensuite un pompier qui ira sauver le participant.

Les Drones peuvent communiquer entre eux pour remonter l'information à la tente de secours. Il faut être à portée de communication de la tente de secours pour transmettre une information.

[Source: [5], [6], [7]]

## 4. Implémentation 

Les Agents utilisent une boucle de Perception/Délibération/Action, et évoluent en parallèle avec des goroutines pour permettre une évolution indépendante et non-déterministe dans la mesure des fonctionnalités du langage go.  
Il a été choisi de synchroniser les agents pour ne leur permettre qu'une itération de leur cycle de perception/délibération/action par tick de la simulation globale pour conserver une cohérence des actions des agents entre eux, et rester plus fidèle aux conditions réelles.

Un objet Simulation contient l'ensemble des éléments utiles à notre simulation, dont une instance de Carte, qui mémorise et gère les positions et déplacements des agents.

Pour l'interface graphique l'outil Ebiten a été utilisé, pour permettre une implémentation globale 100% en Go.

Les images utilisées ont été générées par des IA génératives, puis retouchées ensuite à la main.

[Source: [4], [8]]

## 5. Métriques et Évaluation

### 5.1 Métriques en Temps Réel

Au cours de la simulation sont calculées et affichées quelques informations pour permettre de juger de l'état du système en temps réel :  
- Le nombre de participants  
- Combien sont en situation de détresse  
- Combien ont été traités  
- Combien n'ont pas été pris en charge à temps  
- La batterie moyenne des drones de la flotte  
- La proportion de la surface totale de terrain observée  

### 5.2 Calcul de Performance

Pour évaluer les performances de la flotte de drone, une fois la simulation terminée deux graphiques sont également générés et sauvegardés.  
Le premier graphique représente l'évolution du nombre de personnes en situation de détresse, ainsi que les moments de prise en charge des personnes en fonction du temps.  
Le second graphique représente pour chaque personne sauvée, le temps pris pour le sauvetage. On a ainsi une estimation du temps nécessaire entre le début d'un malaise et l'arrivée d'un secouriste auprès du participant, pour chaque protocole de drone.

## 6. Bibliographie

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
