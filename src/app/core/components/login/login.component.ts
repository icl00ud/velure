import { Component } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';

import { FormControl, FormGroup, NonNullableFormBuilder, Validators } from '@angular/forms';
import { ReactiveFormsModule } from '@angular/forms';

import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzCheckboxModule } from 'ng-zorro-antd/checkbox';
import { TranslateModule } from '@ngx-translate/core';

import { BlobService } from '../../services/blob.service';
import { DomSanitizer, SafeUrl } from '@angular/platform-browser';

@Component({
  selector: 'app-login',
  standalone: true,
  templateUrl: './login.component.html',
  styleUrl: './login.component.less',
  imports: [
    NzFormModule,
    NzInputModule,
    NzButtonModule,
    NzCheckboxModule,
    ReactiveFormsModule,
    TranslateModule,
  ]
})
export class LoginComponent {
  [key: string]: any;
  public passwordErrorTip: string = '';
  public passwordPlaceholder: string = '';
  public userPlaceholder: string = '';
  public userErrorTip: string = '';
  public safeLogoImageUrl: SafeUrl = '';
  public validateForm: FormGroup<{
    userName: FormControl<string>;
    password: FormControl<string>;
    remember: FormControl<boolean>;
  }> = this.fb.group({
    userName: ['', [Validators.required]],
    password: ['', [Validators.required]],
    remember: [true]
  });

  private translations = {
    'LOGIN.USERNAME': 'userPlaceholder',
    'LOGIN.PASSWORD': 'passwordPlaceholder',
    'LOGIN.USERNAME_REQUIRED': 'userErrorTip',
    'LOGIN.PASSWORD_REQUIRED': 'passwordErrorTip'
  };
  private logoUrl: string = '../../../../assets/images/logo-black.png';

  constructor(
    private fb: NonNullableFormBuilder,
    private translateService: TranslateService,
    private blobService: BlobService,
    private sanitizer: DomSanitizer
  ) { }

  ngOnInit() {
    this.blobService.getBase64FromUrl(this.logoUrl).subscribe((base64String: string) => this.safeLogoImageUrl = this.sanitizer.bypassSecurityTrustUrl('data:image/png;base64,' + base64String));

    Object.entries(this.translations).forEach(([key, value]) => {
      this.translateService.get(key).subscribe((res: string) => {
        this[value] = res;
      });
    });
  }

  submitForm(): void {
    if (this.validateForm.valid) {
    } else {
      Object.values(this.validateForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
    }
  }
}
