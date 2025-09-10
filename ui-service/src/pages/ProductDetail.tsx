import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { 
  Heart, 
  Star, 
  Minus, 
  Plus, 
  ShoppingCart, 
  Share2, 
  ArrowLeft,
  Truck,
  Shield,
  RotateCcw,
  Award
} from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { toast } from "@/hooks/use-toast";
import Header from "@/components/Header";

// Mock product data
const mockProduct = {
  id: 1,
  name: "Premium Dry Dog Food - Chicken & Rice",
  brand: "PetNutrition Pro",
  price: 45.99,
  originalPrice: 52.99,
  rating: 4.8,
  reviews: 156,
  category: "Food",
  tags: ["Premium", "Grain-Free", "Adult"],
  inStock: true,
  stockQuantity: 23,
  discount: 13,
  description: "Give your dog the nutrition they deserve with our premium chicken and rice formula. Made with real chicken as the first ingredient, this recipe provides complete and balanced nutrition for adult dogs.",
  features: [
    "Real chicken as first ingredient",
    "No artificial colors, flavors, or preservatives", 
    "Rich in protein for lean muscle maintenance",
    "Added vitamins and minerals for immune support",
    "Omega-6 fatty acids for healthy skin and coat"
  ],
  specifications: {
    "Weight": "15 lbs (6.8 kg)",
    "Life Stage": "Adult",
    "Breed Size": "All Sizes",
    "Primary Protein": "Chicken",
    "Special Diet": "Grain-Free"
  },
  ingredients: "Deboned Chicken, Chicken Meal, Sweet Potatoes, Peas, Chicken Fat, Tomato Pomace, Natural Flavor, Salt, Choline Chloride, Taurine, Dried Chicory Root, Yucca Schidigera Extract, Rosemary Extract, Mixed Tocopherols",
  images: [
    "/api/placeholder/500/500",
    "/api/placeholder/500/500", 
    "/api/placeholder/500/500",
    "/api/placeholder/500/500"
  ]
};

const mockReviews = [
  {
    id: 1,
    author: "Sarah M.",
    rating: 5,
    date: "2024-01-15",
    title: "My dog loves it!",
    content: "My golden retriever absolutely loves this food. His coat is shinier and he has more energy. Highly recommend!"
  },
  {
    id: 2, 
    author: "Mike R.",
    rating: 4,
    date: "2024-01-10",
    title: "Great quality",
    content: "Good quality food at a reasonable price. My dog took a few days to adjust but now eats it happily."
  },
  {
    id: 3,
    author: "Lisa K.",
    rating: 5,
    date: "2024-01-05", 
    title: "Excellent nutrition",
    content: "Vet recommended this brand. Great ingredients and my dog's digestion has improved significantly."
  }
];

const ProductDetail = () => {
  const { id } = useParams();
  const [selectedImage, setSelectedImage] = useState(0);
  const [quantity, setQuantity] = useState(1);
  const [isFavorite, setIsFavorite] = useState(false);

  const handleAddToCart = () => {
    toast({
      title: "Added to cart!",
      description: `${quantity} x ${mockProduct.name} added to your cart.`,
    });
  };

  const handleBuyNow = () => {
    toast({
      title: "Redirecting to checkout...",
      description: "Taking you to secure checkout.",
    });
  };

  return (
    <div className="min-h-screen bg-background">
      <Header />
      
      <main className="container mx-auto px-4 py-8">
        {/* Breadcrumb */}
        <nav className="mb-6">
          <div className="flex items-center space-x-2 text-sm text-muted-foreground">
            <Link to="/" className="hover:text-primary">Home</Link>
            <span>/</span>
            <Link to="/products/dogs" className="hover:text-primary">Dogs</Link>
            <span>/</span>
            <span className="text-foreground font-medium">{mockProduct.category}</span>
          </div>
        </nav>

        <Link
          to="/products/dogs"
          className="inline-flex items-center text-muted-foreground hover:text-primary transition-colors mb-6"
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Products
        </Link>

        <div className="grid lg:grid-cols-2 gap-12 mb-12">
          {/* Product Images */}
          <div className="space-y-4">
            <div className="aspect-square bg-muted rounded-lg flex items-center justify-center text-8xl overflow-hidden">
              🐕
            </div>
            
            <div className="grid grid-cols-4 gap-2">
              {[0, 1, 2, 3].map((index) => (
                <button
                  key={index}
                  onClick={() => setSelectedImage(index)}
                  className={`aspect-square bg-muted rounded-lg flex items-center justify-center text-2xl border-2 transition-colors ${
                    selectedImage === index ? 'border-primary' : 'border-transparent hover:border-muted-foreground'
                  }`}
                >
                  🐕
                </button>
              ))}
            </div>
          </div>

          {/* Product Info */}
          <div className="space-y-6">
            <div>
              <p className="text-muted-foreground font-medium mb-2">{mockProduct.brand}</p>
              <h1 className="text-3xl font-bold text-foreground mb-4">{mockProduct.name}</h1>
              
              <div className="flex items-center space-x-4 mb-4">
                <div className="flex items-center">
                  {[1, 2, 3, 4, 5].map((star) => (
                    <Star 
                      key={star}
                      className={`h-4 w-4 ${
                        star <= Math.floor(mockProduct.rating) 
                          ? 'text-accent fill-current' 
                          : 'text-muted-foreground'
                      }`}
                    />
                  ))}
                  <span className="ml-2 font-medium">{mockProduct.rating}</span>
                </div>
                <span className="text-muted-foreground">({mockProduct.reviews} reviews)</span>
              </div>

              <div className="flex flex-wrap gap-2 mb-6">
                {mockProduct.tags.map((tag, index) => (
                  <Badge key={index} variant="secondary">{tag}</Badge>
                ))}
              </div>
            </div>

            <div className="flex items-baseline space-x-4">
              <span className="text-3xl font-bold text-primary">${mockProduct.price}</span>
              {mockProduct.originalPrice && (
                <span className="text-xl text-muted-foreground line-through">
                  ${mockProduct.originalPrice}
                </span>
              )}
              {mockProduct.discount > 0 && (
                <Badge className="bg-secondary text-secondary-foreground">
                  Save {mockProduct.discount}%
                </Badge>
              )}
            </div>

            <p className="text-muted-foreground leading-relaxed">
              {mockProduct.description}
            </p>

            {/* Stock Status */}
            <div className="flex items-center space-x-2">
              <div className="w-2 h-2 bg-primary rounded-full"></div>
              <span className="text-sm font-medium text-primary">
                In Stock ({mockProduct.stockQuantity} available)
              </span>
            </div>

            {/* Quantity & Actions */}
            <div className="space-y-4">
              <div className="flex items-center space-x-4">
                <span className="font-medium">Quantity:</span>
                <div className="flex items-center space-x-2">
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => setQuantity(Math.max(1, quantity - 1))}
                    disabled={quantity <= 1}
                  >
                    <Minus className="h-4 w-4" />
                  </Button>
                  <span className="w-12 text-center font-medium">{quantity}</span>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => setQuantity(Math.min(mockProduct.stockQuantity, quantity + 1))}
                    disabled={quantity >= mockProduct.stockQuantity}
                  >
                    <Plus className="h-4 w-4" />
                  </Button>
                </div>
              </div>

              <div className="flex space-x-4">
                <Button
                  onClick={handleAddToCart}
                  className="flex-1 bg-gradient-primary hover:opacity-90 text-primary-foreground"
                >
                  <ShoppingCart className="h-4 w-4 mr-2" />
                  Add to Cart
                </Button>
                
                <Button
                  variant="outline"
                  onClick={handleBuyNow}
                  className="flex-1 border-primary text-primary hover:bg-primary hover:text-primary-foreground"
                >
                  Buy Now
                </Button>
              </div>

              <div className="flex space-x-2">
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => setIsFavorite(!isFavorite)}
                  className={isFavorite ? "text-red-500 border-red-500" : ""}
                >
                  <Heart className={`h-4 w-4 ${isFavorite ? 'fill-current' : ''}`} />
                </Button>
                <Button variant="outline" size="icon">
                  <Share2 className="h-4 w-4" />
                </Button>
              </div>
            </div>

            {/* Benefits */}
            <div className="grid grid-cols-2 gap-4 pt-6">
              <div className="flex items-center space-x-3">
                <div className="bg-primary/10 rounded-full p-2">
                  <Truck className="h-4 w-4 text-primary" />
                </div>
                <div>
                  <p className="text-sm font-medium">Free Shipping</p>
                  <p className="text-xs text-muted-foreground">On orders over $50</p>
                </div>
              </div>
              
              <div className="flex items-center space-x-3">
                <div className="bg-secondary/10 rounded-full p-2">
                  <RotateCcw className="h-4 w-4 text-secondary" />
                </div>
                <div>
                  <p className="text-sm font-medium">30-Day Returns</p>
                  <p className="text-xs text-muted-foreground">Money back guarantee</p>
                </div>
              </div>
              
              <div className="flex items-center space-x-3">
                <div className="bg-accent/10 rounded-full p-2">
                  <Shield className="h-4 w-4 text-accent-foreground" />
                </div>
                <div>
                  <p className="text-sm font-medium">Quality Guarantee</p>
                  <p className="text-xs text-muted-foreground">Premium products</p>
                </div>
              </div>
              
              <div className="flex items-center space-x-3">
                <div className="bg-primary/10 rounded-full p-2">
                  <Award className="h-4 w-4 text-primary" />
                </div>
                <div>
                  <p className="text-sm font-medium">Vet Approved</p>
                  <p className="text-xs text-muted-foreground">Trusted by vets</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Product Details Tabs */}
        <Card className="shadow-soft">
          <CardContent className="p-0">
            <Tabs defaultValue="description" className="w-full">
              <TabsList className="grid w-full grid-cols-4">
                <TabsTrigger value="description">Description</TabsTrigger>
                <TabsTrigger value="specifications">Specifications</TabsTrigger>
                <TabsTrigger value="ingredients">Ingredients</TabsTrigger>
                <TabsTrigger value="reviews">Reviews ({mockProduct.reviews})</TabsTrigger>
              </TabsList>
              
              <TabsContent value="description" className="p-6">
                <div className="space-y-4">
                  <h3 className="text-xl font-semibold text-foreground">Product Description</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    {mockProduct.description}
                  </p>
                  
                  <h4 className="text-lg font-semibold text-foreground mt-6">Key Features</h4>
                  <ul className="space-y-2">
                    {mockProduct.features.map((feature, index) => (
                      <li key={index} className="flex items-start space-x-2">
                        <div className="w-1.5 h-1.5 bg-primary rounded-full mt-2 flex-shrink-0"></div>
                        <span className="text-muted-foreground">{feature}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </TabsContent>
              
              <TabsContent value="specifications" className="p-6">
                <div className="space-y-4">
                  <h3 className="text-xl font-semibold text-foreground">Specifications</h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {Object.entries(mockProduct.specifications).map(([key, value]) => (
                      <div key={key} className="flex justify-between py-2 border-b border-border">
                        <span className="font-medium text-foreground">{key}:</span>
                        <span className="text-muted-foreground">{value}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </TabsContent>
              
              <TabsContent value="ingredients" className="p-6">
                <div className="space-y-4">
                  <h3 className="text-xl font-semibold text-foreground">Ingredients</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    {mockProduct.ingredients}
                  </p>
                </div>
              </TabsContent>
              
              <TabsContent value="reviews" className="p-6">
                <div className="space-y-6">
                  <div className="flex items-center justify-between">
                    <h3 className="text-xl font-semibold text-foreground">Customer Reviews</h3>
                    <div className="flex items-center space-x-2">
                      <div className="flex items-center">
                        {[1, 2, 3, 4, 5].map((star) => (
                          <Star 
                            key={star}
                            className={`h-4 w-4 ${
                              star <= Math.floor(mockProduct.rating) 
                                ? 'text-accent fill-current' 
                                : 'text-muted-foreground'
                            }`}
                          />
                        ))}
                      </div>
                      <span className="font-medium">{mockProduct.rating} out of 5</span>
                    </div>
                  </div>
                  
                  <Separator />
                  
                  <div className="space-y-6">
                    {mockReviews.map((review) => (
                      <div key={review.id} className="space-y-3">
                        <div className="flex items-start justify-between">
                          <div>
                            <div className="flex items-center space-x-2 mb-1">
                              <span className="font-medium text-foreground">{review.author}</span>
                              <div className="flex items-center">
                                {[1, 2, 3, 4, 5].map((star) => (
                                  <Star 
                                    key={star}
                                    className={`h-3 w-3 ${
                                      star <= review.rating 
                                        ? 'text-accent fill-current' 
                                        : 'text-muted-foreground'
                                    }`}
                                  />
                                ))}
                              </div>
                            </div>
                            <h4 className="font-medium text-foreground">{review.title}</h4>
                          </div>
                          <span className="text-sm text-muted-foreground">{review.date}</span>
                        </div>
                        <p className="text-muted-foreground">{review.content}</p>
                        <Separator />
                      </div>
                    ))}
                  </div>
                </div>
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </main>
    </div>
  );
};

export default ProductDetail;