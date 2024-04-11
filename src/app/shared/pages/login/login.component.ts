import { Component } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';

import { FormControl, FormGroup, NonNullableFormBuilder, Validators } from '@angular/forms';
import { ReactiveFormsModule } from '@angular/forms';

import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzCheckboxModule } from 'ng-zorro-antd/checkbox';
import { TranslateModule } from '@ngx-translate/core';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    NzFormModule,
    NzInputModule,
    NzButtonModule,
    NzCheckboxModule,
    ReactiveFormsModule,
    TranslateModule
  ],
  templateUrl: './login.component.html',
  styleUrl: './login.component.less'
})
export class LoginComponent {
  [key: string]: any;
  public passwordErrorTip: string = '';
  public passwordPlaceholder: string = '';
  public userPlaceholder: string = '';
  public userErrorTip: string = '';

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

  constructor(
    private fb: NonNullableFormBuilder,
    private translateService: TranslateService,
  ) { }

  ngOnInit() {
    Object.entries(this.translations).forEach(([key, value]) => {
      this.translateService.get(key).subscribe((res: string) => {
        this[value] = res;
      });
    });
  }

  submitForm(): void {
    if (this.validateForm.valid) {
      console.log('submit', this.validateForm.value);
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
