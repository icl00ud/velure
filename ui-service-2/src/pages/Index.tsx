import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Heart, Shield, Truck, Users } from "lucide-react";
import { Link } from "react-router-dom";
import Header from "@/components/Header";
import heroImage from "@/assets/petshop-hero.jpg";

const Index = () => {
  return (
    <div className="min-h-screen bg-background">
      <Header />
      
      {/* Hero Section */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0">
          <img
            src={heroImage}
            alt="Happy pets in our store"
            className="w-full h-full object-cover"
          />
          <div className="absolute inset-0 bg-gradient-to-r from-background/90 via-background/50 to-transparent"></div>
        </div>
        
        <div className="relative container mx-auto px-4 py-24 lg:py-32">
          <div className="max-w-2xl">
            <h1 className="text-5xl lg:text-6xl font-bold text-foreground mb-6">
              Everything Your
              <span className="text-primary block">Pet Needs</span>
            </h1>
            <p className="text-xl text-muted-foreground mb-8 leading-relaxed">
              From premium food to cozy beds, toys to health care - we have everything 
              to keep your furry, feathered, and finned friends happy and healthy.
            </p>
            <div className="flex flex-col sm:flex-row gap-4">
              <Button 
                size="lg" 
                className="bg-gradient-primary hover:opacity-90 text-primary-foreground px-8"
                asChild
              >
                <Link to="/products/dogs">Shop Now</Link>
              </Button>
              <Button 
                variant="outline" 
                size="lg" 
                className="border-primary text-primary hover:bg-primary hover:text-primary-foreground"
                asChild
              >
                <Link to="/contact">Learn More</Link>
              </Button>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 bg-muted/30">
        <div className="container mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="text-3xl lg:text-4xl font-bold text-foreground mb-4">
              Why Pet Parents Choose Us
            </h2>
            <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
              We're more than just a pet store - we're your partner in pet care
            </p>
          </div>
          
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
            <Card className="text-center shadow-soft hover:shadow-primary transition-all duration-300">
              <CardHeader>
                <div className="mx-auto bg-gradient-primary rounded-full p-4 w-16 h-16 flex items-center justify-center mb-4">
                  <Heart className="h-8 w-8 text-primary-foreground" />
                </div>
                <CardTitle className="text-primary">Premium Quality</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  Only the best products from trusted brands for your beloved pets
                </CardDescription>
              </CardContent>
            </Card>

            <Card className="text-center shadow-soft hover:shadow-secondary transition-all duration-300">
              <CardHeader>
                <div className="mx-auto bg-gradient-secondary rounded-full p-4 w-16 h-16 flex items-center justify-center mb-4">
                  <Shield className="h-8 w-8 text-secondary-foreground" />
                </div>
                <CardTitle className="text-secondary">Health Guaranteed</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  All our products are vet-approved and come with health guarantees
                </CardDescription>
              </CardContent>
            </Card>

            <Card className="text-center shadow-soft hover:shadow-accent transition-all duration-300">
              <CardHeader>
                <div className="mx-auto bg-gradient-accent rounded-full p-4 w-16 h-16 flex items-center justify-center mb-4">
                  <Truck className="h-8 w-8 text-accent-foreground" />
                </div>
                <CardTitle className="text-accent-foreground">Fast Delivery</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  Free shipping on orders over $50. Same-day delivery available
                </CardDescription>
              </CardContent>
            </Card>

            <Card className="text-center shadow-soft hover:shadow-primary transition-all duration-300">
              <CardHeader>
                <div className="mx-auto bg-gradient-primary rounded-full p-4 w-16 h-16 flex items-center justify-center mb-4">
                  <Users className="h-8 w-8 text-primary-foreground" />
                </div>
                <CardTitle className="text-primary">Expert Support</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  Our pet care experts are here to help you make the best choices
                </CardDescription>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>

      {/* Popular Categories */}
      <section className="py-20">
        <div className="container mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="text-3xl lg:text-4xl font-bold text-foreground mb-4">
              Shop by Pet
            </h2>
            <p className="text-lg text-muted-foreground">
              Find everything your specific pet needs
            </p>
          </div>
          
          <div className="grid sm:grid-cols-2 lg:grid-cols-5 gap-6">
            {[
              { name: "Dogs", emoji: "🐕", color: "primary", link: "/products/dogs" },
              { name: "Cats", emoji: "🐱", color: "secondary", link: "/products/cats" },
              { name: "Birds", emoji: "🦜", color: "accent", link: "/products/birds" },
              { name: "Fish", emoji: "🐠", color: "primary", link: "/products/fish" },
              { name: "Small Pets", emoji: "🐹", color: "secondary", link: "/products/small-pets" },
            ].map((category) => (
              <Link key={category.name} to={category.link}>
                <Card className="group cursor-pointer hover:scale-105 transition-transform duration-300 shadow-soft">
                  <CardContent className="p-8 text-center">
                    <div className="text-6xl mb-4">{category.emoji}</div>
                    <h3 className="font-semibold text-lg text-foreground group-hover:text-primary transition-colors">
                      {category.name}
                    </h3>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 bg-gradient-hero">
        <div className="container mx-auto px-4 text-center">
          <div className="max-w-3xl mx-auto">
            <h2 className="text-4xl lg:text-5xl font-bold text-primary-foreground mb-6">
              Ready to Spoil Your Pet?
            </h2>
            <p className="text-xl text-primary-foreground/90 mb-8 leading-relaxed">
              Join thousands of happy pet parents who trust us with their pets' happiness and health
            </p>
            <Button 
              size="lg" 
              className="bg-background text-primary hover:bg-background/90 px-8"
              asChild
            >
              <Link to="/products/dogs">Start Shopping</Link>
            </Button>
          </div>
        </div>
      </section>
    </div>
  );
};

export default Index;