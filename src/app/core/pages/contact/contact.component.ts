import { Component } from '@angular/core';

import { FormsModule } from '@angular/forms';

declare var google: any;

@Component({
  selector: 'app-contact',
  standalone: true,
  imports: [
    FormsModule
  ],
  templateUrl: './contact.component.html',
  styleUrl: './contact.component.less'
})
export class ContactComponent {
  constructor() { }

  ngOnInit(): void {
    this.loadMap();
  }

  loadMap(): void {
    const script = document.createElement('script');
    script.src = `https://maps.googleapis.com/maps/api/js?key=AIzaSyCy7z0_R4tnPvrBmpZOXi9vvSaWzSKj1rM&callback=initMap`;
    script.async = true;
    script.defer = true;
    document.head.appendChild(script);

    (window as any).initMap = () => {
      const map = new google.maps.Map(document.getElementById('map') as HTMLElement, {
        center: { lat: -26.3044, lng: -48.8487 },
        zoom: 15,
      });
    };
  }
}
