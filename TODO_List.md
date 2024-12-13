# TODO List IA04

## Backend

- [ ] Mettre en place une répartition uniforme des drones dans la map.  
- [ ] Calcul des statistiques de la simulation.  
- [ ] **Protocole 2** :  
  - [ ] Les drones propagent de proche en proche l’information d’une personne `InDistress`.  
  - [ ] Le drone le plus proche du `medical gear` ramène l’équipement.  
- [ ] **Protocole 3** :  
  - [ ] Les drones propagent de proche en proche l’information d’une personne `InDistress`.  
  - [ ] Un bonhomme sauveteur sort du `medical gear` pour aller sauver la personne.  
- [ ] **Protocole 4** :  
  - [ ] Mise en place d’un système de leader pour chaque groupe de drones.  
  - [ ] Un drone va chercher le `medical gear`.  
- [ ] **Protocole 5** :  
  - [ ] Mise en place d’un système de leader pour chaque groupe de drones.  
  - [ ] Un bonhomme sort du `medical gear` pour aller sauver la personne.  

## Frontend

- [ ] Affichage des statistiques de la simulation.  
- [ ] Correction pour aligner la sélection du nombre de personnes/drones dans le menu.  
- [ ] Afficher les bonhommes sauveteurs d’une autre couleur.  
- [ ] Afficher d’une couleur différente les personnes détectées `InDistress` (en cours de sauvetage). 
- [ ] Menu déroulant pour le choix du protocole lors du lancement de la simulation

---

## Finalization Check

```python
if allITemInListChecked():
    return "End of project!"
else:
    return "Continue working."
```