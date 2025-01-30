import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'currency',
  standalone: true
})
export class CurrencyPipe implements PipeTransform {
  transform(
    value: number,
    currencySymbol: string = 'R$',
    decimalPlaces: number = 2
  ): string {
    if (value == null) {
      return `${currencySymbol} 0.00`;
    }

    const formattedValue = value.toFixed(decimalPlaces).replace(/\d(?=(\d{3})+\.)/g, '$&,');
    return `${currencySymbol} ${formattedValue}`;
  }
}
