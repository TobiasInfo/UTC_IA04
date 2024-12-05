package assets

// POIIcon returns an SVG string for the given POI type
func POIIcon(poiType int) string {
	switch poiType {
	case 0: // MedicalTent
		return `<svg viewBox="0 0 24 24">
			<path fill="red" d="M19 3H5C3.89 3 3 3.89 3 5V19C3 20.11 3.89 21 5 21H19C20.11 21 21 20.11 21 19V5C21 3.89 20.11 3 19 3M18 14H14V18H10V14H6V10H10V6H14V10H18V14Z"/>
		</svg>`
	case 1: // ChargingStation
		return `<svg viewBox="0 0 24 24">
			<path fill="yellow" d="M12 3L2 12H5V20H19V12H22L12 3M11 10V18H8V12.5L5 14.5L12 7L19 14.5L16 12.5V18H13V10H11Z"/>
		</svg>`
	case 2: // Toilet
		return `<svg viewBox="0 0 24 24">
			<path fill="gray" d="M5.5 22V12.5H10.5V22H5.5M17.5 22V12.5H22.5V22H17.5M11.5 22V3.5H16.5V22H11.5Z"/>
		</svg>`
	case 3: // DrinkStand
		return `<svg viewBox="0 0 24 24" width="16" height="16">
    <path fill="skyblue" d="M3 2L5 20.23C5.13 21.23 5.97 22 7 22H17C18 22 18.87 21.23 19 20.23L21 2H3M12 19C10.9 19 10 18.1 10 17C10 15.9 10.9 15 12 15C13.1 15 14 15.9 14 17C14 18.1 13.1 19 12 19Z"/>
</svg>`
	case 4: // FoodStand
		return `<svg viewBox="0 0 24 24">
			<path fill="orange" d="M15.5 21L14 8H16.23L15.1 3.46L16.84 3L18.09 8H22L20.5 21H15.5M5 11H10C11.66 11 13 12.34 13 14H2C2 12.34 3.34 11 5 11M13 18H2V15H13V18Z"/>
		</svg>`
	case 5: // MainStage
		return `<svg viewBox="0 0 24 24">
			<path fill="purple" d="M22 4V2H2V4H11V18.5C11 19.88 9.88 21 8.5 21S6 19.88 6 18.5H4C4 20.99 6.01 23 8.5 23S13 20.99 13 18.5V4H22Z"/>
		</svg>`
	case 6: // SecondaryStage
		return `<svg viewBox="0 0 24 24">
			<path fill="mediumpurple" d="M19 3H5C3.89 3 3 3.89 3 5V19C3 20.1 3.89 21 5 21H19C20.1 21 21 20.1 21 19V5C21 3.89 20.11 3 19 3M19 19H5V5H19V19Z"/>
		</svg>`
	case 7: // RestArea
		return `<svg viewBox="0 0 24 24">
			<path fill="seagreen" d="M4 18V21H7V18H4M9 18V21H12V18H9M14 18V21H17V18H14M4 13V16H7V13H4M9 13V16H12V13H9M14 13V16H17V13H14M4 8V11H7V8H4M9 8V11H12V8H9M14 8V11H17V8H14Z"/>
		</svg>`
	default:
		return ""
	}
}
