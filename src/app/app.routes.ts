import { Routes } from '@angular/router';

import { HomeComponent } from './core/pages/home/home.component';
import { PageNotFoundComponent } from './core/pages/page-not-found/page-not-found.component';
import { LoginComponent } from './core/pages/login/login.component';
import { CategoryProductsComponent } from './core/pages/category-products/category-products.component';
import { RegisterComponent } from './core/pages/register/register.component';
import { ForgotPasswordComponent } from './core/pages/forgot-password/forgot-password.component';
import { ContactComponent } from './core/pages/contact/contact.component';

import { AuthGuard } from './core/guards/auth.guard';
import { NoAuthGuard } from './core/guards/noAuth.guard';

const categoryData = [
  { path: 'shoes', data: { category: 'shoes' } },
  { path: 'shirts', data: { category: 'shirts' } },
];

export const routes: Routes = [
  { path: '', redirectTo: 'home', pathMatch: 'full', canActivate: [AuthGuard] },
  { path: 'home', component: HomeComponent, canActivate: [AuthGuard] },
  { path: 'login', component: LoginComponent, canActivate: [NoAuthGuard] },
  { path: 'register', component: RegisterComponent, canActivate: [NoAuthGuard] },
  { path: 'forgot-password', component: ForgotPasswordComponent },
  { path: 'contact', component: ContactComponent},
  ...categoryData.map((category) => ({
    ...category,
    component: CategoryProductsComponent,
    canActivate: [AuthGuard],
  })),
  { path: '**', component: PageNotFoundComponent },
];